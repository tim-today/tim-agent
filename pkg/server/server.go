package server

import (
	"crypto/subtle"
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/skip2/go-qrcode"

	"github.com/tim-today/tim-agent/pkg/config"
	"github.com/tim-today/tim-agent/pkg/network"
	"github.com/tim-today/tim-agent/pkg/session"
	"github.com/tim-today/tim-agent/pkg/web"
)

type Server struct {
	cfg        *config.Config
	tmpl       *template.Template
	upgrader   websocket.Upgrader
	srv        *http.Server // 外部共享终端服务 (0.0.0.0:cfg.Port)
	adminSrv   *http.Server // 本地免密控制面板 (127.0.0.1:adminPort 随机端口)
	adminPort  int
	publicPort int
	mu         sync.RWMutex
	startTime  time.Time
}

func recoveryMiddleware(name string, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if rec := recover(); rec != nil {
				log.Printf("[TimAgent 异常守护] %s 捕获未处理 panic: %v", name, rec)
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			}
		}()
		next.ServeHTTP(w, r)
	})
}

func NewServer(cfg *config.Config) (*Server, error) {
	tmpl, err := web.GetTemplates()
	if err != nil {
		return nil, fmt.Errorf("加载网页模板失败: %w", err)
	}

	s := &Server{
		cfg:  cfg,
		tmpl: tmpl,
		upgrader: websocket.Upgrader{
			ReadBufferSize:  4096,
			WriteBufferSize: 4096,
			CheckOrigin: func(r *http.Request) bool {
				return true // 允许跨源与局域网任意访问
			},
		},
		startTime: time.Now(),
	}

	return s, nil
}

func (s *Server) AdminPort() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.adminPort
}

func (s *Server) PublicPort() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.publicPort
}

func (s *Server) Start() error {
	// 1. 启动本地控制面板服务 (默认绑定 127.0.0.1:20997, 若占用则自动递增)
	var adminLn net.Listener
	var err error
	startAdminPort := 20997
	for port := startAdminPort; port < startAdminPort+100; port++ {
		adminLn, err = net.Listen("tcp", fmt.Sprintf("127.0.0.1:%d", port))
		if err == nil {
			break
		}
	}
	if adminLn == nil {
		// 若范围内均被占用，回退到系统随机端口
		adminLn, err = net.Listen("tcp", "127.0.0.1:0")
		if err != nil {
			return fmt.Errorf("绑定本地控制面板端口失败: %w", err)
		}
	}

	s.mu.Lock()
	s.adminPort = adminLn.Addr().(*net.TCPAddr).Port
	s.mu.Unlock()

	adminMux := http.NewServeMux()
	adminMux.Handle("/css/", http.FileServer(web.GetStaticFS()))
	adminMux.Handle("/js/", http.FileServer(web.GetStaticFS()))
	adminMux.HandleFunc("/", s.handleDashboard) // 本机直开控制面板，无需密码
	adminMux.HandleFunc("/terminal", s.handleTerminal)
	adminMux.HandleFunc("/api/status", s.handleAPIStatus)
	adminMux.HandleFunc("/api/qrcode", s.handleAPIQRCode)
	adminMux.HandleFunc("/api/config", s.handleAPIConfig)
	adminMux.HandleFunc("/api/password/regenerate", s.handleAPIRegeneratePassword)
	adminMux.HandleFunc("/api/sessions/restart", s.handleAPISessionRestart)
	adminMux.HandleFunc("/api/version/check", s.handleAPIVersionCheck)
	adminMux.HandleFunc("/ws/terminal", s.handleWSTerminal)

	s.adminSrv = &http.Server{
		Handler: recoveryMiddleware("AdminPanel", adminMux),
	}

	go func() {
		if err := s.adminSrv.Serve(adminLn); err != nil && err != http.ErrServerClosed {
			log.Printf("[TimAgent] 本地控制面板服务异常: %v", err)
		}
	}()
	log.Printf("[TimAgent] 本地安全控制面板已就绪: http://127.0.0.1:%d/ (本机专属, 免密码)", s.adminPort)

	// 2. 启动外部共享终端服务 (绑定 0.0.0.0:cfg.Port, 若占用则自动递增)
	publicMux := http.NewServeMux()
	publicMux.Handle("/css/", http.FileServer(web.GetStaticFS()))
	publicMux.Handle("/js/", http.FileServer(web.GetStaticFS()))

	// 外部访问根路径自动跳转到终端登录页，防止外部探测或越权管理
	publicMux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}
		http.Redirect(w, r, "/terminal", http.StatusFound)
	})

	publicMux.HandleFunc("/terminal", s.handleTerminal)
	publicMux.HandleFunc("/login", s.handleLogin)
	publicMux.HandleFunc("/api/login", s.handleAPILogin)
	publicMux.HandleFunc("/api/status", s.handleAPIStatus)
	publicMux.HandleFunc("/api/qrcode", s.handleAPIQRCode)
	publicMux.HandleFunc("/ws/terminal", s.handleWSTerminal)

	cfg := s.cfg.Get()
	targetPort := cfg.Port
	if targetPort <= 0 {
		targetPort = 20996
	}

	var pubLn net.Listener
	for port := targetPort; port < targetPort+100; port++ {
		// 跳过已被本地管理端占用的端口
		if port == s.adminPort {
			continue
		}
		pubLn, err = net.Listen("tcp", fmt.Sprintf("0.0.0.0:%d", port))
		if err == nil {
			if port != cfg.Port {
				// 若递增了端口，同步更新配置中的当前实际端口
				_ = s.cfg.Update(port, cfg.Password, cfg.Shell, cfg.WorkDir, cfg.SelectedAgent, cfg.Theme, cfg.Language, cfg.AutoApprove, cfg.KeepAlive, nil)
			}
			break
		}
	}
	if pubLn == nil {
		return fmt.Errorf("未能成功绑定外部共享端口 (从 %d 开始均被占用): %w", targetPort, err)
	}

	actualPubPort := pubLn.Addr().(*net.TCPAddr).Port
	s.mu.Lock()
	s.publicPort = actualPubPort
	s.mu.Unlock()

	addr := fmt.Sprintf(":%d", actualPubPort)

	s.srv = &http.Server{
		Addr:         addr,
		Handler:      recoveryMiddleware("PublicTerminal", publicMux),
		ReadTimeout:  0, // WebSocket 不设硬超时
		WriteTimeout: 0,
	}

	log.Printf("[TimAgent] 外部终端共享服务已在 %s 启动 (需密码或Token鉴权)", addr)
	return s.srv.Serve(pubLn)
}

func (s *Server) Close() error {
	if s.adminSrv != nil {
		_ = s.adminSrv.Close()
	}
	if s.srv != nil {
		return s.srv.Close()
	}
	return nil
}

// isAuthenticated 鉴权检查
func (s *Server) isAuthenticated(r *http.Request) bool {
	cfg := s.cfg.Get()
	if cfg.Password == "" {
		return true // 未设密码免鉴权
	}

	// 1. URL token 快速免密 (用于扫码)
	token := r.URL.Query().Get("token")
	if token != "" && subtle.ConstantTimeCompare([]byte(token), []byte(cfg.AuthToken)) == 1 {
		return true
	}

	// 2. Cookie 鉴权
	cookie, err := r.Cookie("tim_session")
	if err == nil && cookie != nil {
		if subtle.ConstantTimeCompare([]byte(cookie.Value), []byte(cfg.AuthToken)) == 1 {
			return true
		}
	}

	return false
}

func (s *Server) handleDashboard(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}

	// 本地控制面板限制127.0.0.1随机端口访问，无需输入密码，直接打开控制面板
	if err := s.tmpl.ExecuteTemplate(w, "index.html", nil); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (s *Server) handleTerminal(w http.ResponseWriter, r *http.Request) {
	if !s.isAuthenticated(r) {
		http.Redirect(w, r, "/login?redirect="+r.URL.RequestURI(), http.StatusFound)
		return
	}

	// 如果是通过 token 访问，自动写入 cookie，后续操作免密
	token := r.URL.Query().Get("token")
	cfg := s.cfg.Get()
	if token != "" && token == cfg.AuthToken {
		http.SetCookie(w, &http.Cookie{
			Name:     "tim_session",
			Value:    cfg.AuthToken,
			Path:     "/",
			HttpOnly: true,
			SameSite: http.SameSiteLaxMode,
			MaxAge:   365 * 24 * 3600,
			Expires:  time.Now().Add(365 * 24 * time.Hour),
		})
	}

	if err := s.tmpl.ExecuteTemplate(w, "terminal.html", nil); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (s *Server) handleLogin(w http.ResponseWriter, r *http.Request) {
	if err := s.tmpl.ExecuteTemplate(w, "login.html", nil); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (s *Server) handleAPILogin(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	cfg := s.cfg.Get()
	if cfg.Password == "" || subtle.ConstantTimeCompare([]byte(req.Password), []byte(cfg.Password)) == 1 {
		http.SetCookie(w, &http.Cookie{
			Name:     "tim_session",
			Value:    cfg.AuthToken,
			Path:     "/",
			HttpOnly: true,
			SameSite: http.SameSiteLaxMode,
			MaxAge:   365 * 24 * 3600, // 浏览器持久保存1年
			Expires:  time.Now().Add(365 * 24 * time.Hour),
		})

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{"ok": true})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{"ok": false, "message": "密码错误"})
}

func (s *Server) handleAPIRegeneratePassword(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	newPwd := s.cfg.RegeneratePassword()
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{
		"ok":       true,
		"password": newPwd,
	})
}

func (s *Server) handleAPIStatus(w http.ResponseWriter, r *http.Request) {
	cfg := s.cfg.Get()
	networks := network.GetAvailableIPs(cfg.Port)

	// 获取工作目录绑定的主会话状态
	mgr := session.GetManager()
	initCmd := s.cfg.GetAgentCommand(cfg.SelectedAgent, cfg.AutoApprove)
	defaultSess, _ := mgr.GetOrCreateForDir("default", "主会话", cfg.Shell, cfg.WorkDir, cfg.SelectedAgent, initCmd, cfg.Language)

	var sessInfo map[string]any
	if defaultSess != nil {
		sessInfo = map[string]any{
			"id":            defaultSess.ID,
			"pid":           defaultSess.Pid,
			"work_dir":      defaultSess.WorkDir,
			"active_agent":  defaultSess.ActiveAgent,
			"is_running":    defaultSess.IsRunning,
			"clients_count": len(defaultSess.ListClients()),
		}
	}

	res := map[string]any{
		"version":      config.AppVersion,
		"github_url":   config.GitHubRepoURL,
		"releases_url": config.GitHubReleases,
		"config":       cfg,
		"networks":     networks,
		"auth_token":   cfg.AuthToken,
		"session":      sessInfo,
		"uptime":       time.Since(s.startTime).String(),
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(res)
}

func (s *Server) handleAPIVersionCheck(w http.ResponseWriter, r *http.Request) {
	client := &http.Client{
		Timeout: 5 * time.Second,
	}

	req, err := http.NewRequest("GET", config.GitHubAPIURL, nil)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	req.Header.Set("User-Agent", "Tim-Agent-App/"+config.AppVersion)
	req.Header.Set("Accept", "application/vnd.github.v3+json")

	resp, err := client.Do(req)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"ok":              false,
			"current_version": config.AppVersion,
			"error":           "无法连接至 GitHub 检测版本: " + err.Error(),
		})
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"ok":              false,
			"current_version": config.AppVersion,
			"error":           fmt.Sprintf("GitHub 返回状态码: %d", resp.StatusCode),
		})
		return
	}

	var ghRelease struct {
		TagName string `json:"tag_name"`
		Name    string `json:"name"`
		HTMLURL string `json:"html_url"`
		Body    string `json:"body"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&ghRelease); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	hasUpdate := false
	if ghRelease.TagName != "" && ghRelease.TagName != config.AppVersion {
		hasUpdate = true
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{
		"ok":              true,
		"current_version": config.AppVersion,
		"latest_version":  ghRelease.TagName,
		"has_update":      hasUpdate,
		"html_url":        ghRelease.HTMLURL,
		"release_notes":   ghRelease.Body,
	})
}

func (s *Server) handleAPIQRCode(w http.ResponseWriter, r *http.Request) {
	url := r.URL.Query().Get("url")
	if url == "" {
		http.Error(w, "url required", http.StatusBadRequest)
		return
	}

	pngData, err := qrcode.Encode(url, qrcode.Medium, 240)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "image/png")
	w.Header().Set("Cache-Control", "no-cache")
	_, _ = w.Write(pngData)
}

func (s *Server) handleAPIConfig(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		Port          int    `json:"port"`
		Password      string `json:"password"`
		Shell         string `json:"shell"`
		WorkDir       string `json:"work_dir"`
		SelectedAgent string `json:"selected_agent"`
		Theme         string `json:"theme"`
		Language      string `json:"language"`
		AutoApprove   bool   `json:"auto_approve"`
		KeepAlive     bool   `json:"keep_alive"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	oldCfg := s.cfg.Get()
	agentChanged := (req.SelectedAgent != "" && req.SelectedAgent != oldCfg.SelectedAgent)
	workDirChanged := (req.WorkDir != "" && req.WorkDir != oldCfg.WorkDir)
	autoApproveChanged := (req.AutoApprove != oldCfg.AutoApprove)

	targetLang := req.Language
	if targetLang == "" {
		targetLang = oldCfg.Language
	}

	err := s.cfg.Update(req.Port, req.Password, req.Shell, req.WorkDir, req.SelectedAgent, req.Theme, targetLang, req.AutoApprove, req.KeepAlive, nil)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{"ok": false, "message": err.Error()})
		return
	}

	restarted := false
	if agentChanged || workDirChanged || autoApproveChanged {
		targetWorkDir := req.WorkDir
		if targetWorkDir == "" {
			targetWorkDir = oldCfg.WorkDir
		}
		targetShell := req.Shell
		if targetShell == "" {
			targetShell = oldCfg.Shell
		}
		initCmd := s.cfg.GetAgentCommand(req.SelectedAgent, req.AutoApprove)
		mgr := session.GetManager()
		_, _ = mgr.Restart("default", "主会话", targetShell, targetWorkDir, req.SelectedAgent, initCmd, targetLang)
		restarted = true
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{
		"ok":        true,
		"restarted": restarted,
		"agent":     req.SelectedAgent,
	})
}

func (s *Server) handleAPISessionRestart(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		Agent       string `json:"agent"`
		WorkDir     string `json:"work_dir"`
		AutoApprove *bool  `json:"auto_approve"`
	}
	_ = json.NewDecoder(r.Body).Decode(&req)

	cfg := s.cfg.Get()
	workDir := cfg.WorkDir
	if req.WorkDir != "" {
		workDir = req.WorkDir
	}

	agentID := cfg.SelectedAgent
	if req.Agent != "" {
		agentID = req.Agent
	}

	autoApprove := cfg.AutoApprove
	if req.AutoApprove != nil {
		autoApprove = *req.AutoApprove
	}

	_ = s.cfg.Update(cfg.Port, cfg.Password, cfg.Shell, workDir, agentID, cfg.Theme, cfg.Language, autoApprove, cfg.KeepAlive, nil)

	initCmd := s.cfg.GetAgentCommand(agentID, autoApprove)
	mgr := session.GetManager()
	_, err := mgr.Restart("default", "主会话", cfg.Shell, workDir, agentID, initCmd, cfg.Language)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{"ok": false, "message": err.Error()})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{"ok": true})
}

func (s *Server) handleWSTerminal(w http.ResponseWriter, r *http.Request) {
	if !s.isAuthenticated(r) {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	conn, err := s.upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("[TimAgent] WebSocket Upgrade 失败: %v", err)
		return
	}

	sessID := r.URL.Query().Get("session")
	if sessID == "" {
		sessID = "default"
	}

	cfg := s.cfg.Get()
	initCmd := s.cfg.GetAgentCommand(cfg.SelectedAgent, cfg.AutoApprove)
	sess, err := session.GetManager().GetOrCreateForDir(sessID, "会话-"+sessID, cfg.Shell, cfg.WorkDir, cfg.SelectedAgent, initCmd, cfg.Language)
	if err != nil {
		_ = conn.WriteMessage(websocket.TextMessage, []byte("启动终端失败: "+err.Error()))
		_ = conn.Close()
		return
	}

	// 挂载连接到会话
	sess.Attach(conn)
}
