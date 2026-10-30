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

	"github.com/tim-today/tim-agent/pkg/appbundle"
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

	// 当用户在任一项目会话中进行输入或连接时，自动记录该项目为“最近操作的项目”
	session.GetManager().SetOnActivityCallback(func(sessID string) {
		if sessID != "" && sessID != "default" {
			cfg.SetActiveProject(sessID)
		}
	})

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
	adminMux.Handle("/img/", http.FileServer(web.GetStaticFS()))
	adminMux.HandleFunc("/manifest.json", s.handleManifest)
	adminMux.HandleFunc("/sw.js", s.handleServiceWorker)
	adminMux.HandleFunc("/favicon.ico", s.handleFavicon)
	adminMux.HandleFunc("/", s.handleDashboard) // 本机直开控制面板，无需密码
	adminMux.HandleFunc("/terminal", s.handleTerminal)
	adminMux.HandleFunc("/api/status", s.handleAPIStatus)
	adminMux.HandleFunc("/api/qrcode", s.handleAPIQRCode)
	adminMux.HandleFunc("/api/config", s.handleAPIConfig)
	adminMux.HandleFunc("/api/projects/save", s.handleAPIProjectSave)
	adminMux.HandleFunc("/api/projects/delete", s.handleAPIProjectDelete)
	adminMux.HandleFunc("/api/projects/select", s.handleAPIProjectSelect)
	adminMux.HandleFunc("/api/password/regenerate", s.handleAPIRegeneratePassword)
	adminMux.HandleFunc("/api/sessions/restart", s.handleAPISessionRestart)
	adminMux.HandleFunc("/api/version/check", s.handleAPIVersionCheck)
	adminMux.HandleFunc("/api/app/install", s.handleAPIAppInstall)
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
	publicMux.Handle("/img/", http.FileServer(web.GetStaticFS()))
	publicMux.HandleFunc("/manifest.json", s.handleManifest)
	publicMux.HandleFunc("/sw.js", s.handleServiceWorker)
	publicMux.HandleFunc("/favicon.ico", s.handleFavicon)

	// 外部终端共享服务
	publicMux.HandleFunc("/", s.handleDashboard)
	publicMux.HandleFunc("/terminal", s.handleTerminal)
	publicMux.HandleFunc("/login", s.handleLogin)
	publicMux.HandleFunc("/api/login", s.handleAPILogin)
	publicMux.HandleFunc("/api/status", s.handleAPIStatus)
	publicMux.HandleFunc("/api/qrcode", s.handleAPIQRCode)
	publicMux.HandleFunc("/api/config", s.handleAPIConfig)
	publicMux.HandleFunc("/api/projects/save", s.handleAPIProjectSave)
	publicMux.HandleFunc("/api/projects/delete", s.handleAPIProjectDelete)
	publicMux.HandleFunc("/api/projects/select", s.handleAPIProjectSelect)
	publicMux.HandleFunc("/api/password/regenerate", s.handleAPIRegeneratePassword)
	publicMux.HandleFunc("/api/sessions/restart", s.handleAPISessionRestart)
	publicMux.HandleFunc("/api/version/check", s.handleAPIVersionCheck)
	publicMux.HandleFunc("/api/app/install", s.handleAPIAppInstall)
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
				_ = s.cfg.Update(port, cfg.Password, cfg.Shell, cfg.WorkDir, cfg.SelectedAgent, cfg.Theme, cfg.Language, cfg.AutoApprove, cfg.KeepAlive, cfg.RemoteAccess, nil)
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

// isLocalRequest 判断请求是否来自本机回环地址 (127.0.0.1 / ::1 / localhost)
func isLocalRequest(r *http.Request) bool {
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		host = r.RemoteAddr
	}
	ip := net.ParseIP(host)
	if ip != nil {
		return ip.IsLoopback()
	}
	return host == "localhost" || host == "127.0.0.1" || host == "::1"
}

// requireAdminAuth 远程管理 API 权限校验：
// 1. 本机 127.0.0.1 / localhost 访问直接免密放行；
// 2. 外部 IP 访问：若未开启 remote_access 返回 403 Forbidden；若已开启但未认证返回 401 Unauthorized。
func (s *Server) requireAdminAuth(w http.ResponseWriter, r *http.Request) bool {
	if isLocalRequest(r) {
		return true
	}
	cfg := s.cfg.Get()
	if !cfg.RemoteAccess {
		http.Error(w, "Forbidden: Remote management is disabled", http.StatusForbidden)
		return false
	}
	if !s.isAuthenticated(r) {
		http.Error(w, "Unauthorized: Please login with access password", http.StatusUnauthorized)
		return false
	}
	return true
}

func (s *Server) handleDashboard(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}

	cfg := s.cfg.Get()
	if !isLocalRequest(r) {
		// 外部 IP 请求
		if !cfg.RemoteAccess {
			// 未开启远程管理时，重定向至终端页面
			http.Redirect(w, r, "/terminal", http.StatusFound)
			return
		}
		// 开启了远程管理，必须经过密码鉴权
		if !s.isAuthenticated(r) {
			http.Redirect(w, r, "/login?redirect=/", http.StatusFound)
			return
		}
	}

	// 本机直接免密，或者远程鉴权成功
	if err := s.tmpl.ExecuteTemplate(w, "index.html", nil); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// handleManifest 返回 PWA Web App Manifest
func (s *Server) handleManifest(w http.ResponseWriter, r *http.Request) {
	data, err := web.GetStaticFile("manifest.json")
	if err != nil {
		http.NotFound(w, r)
		return
	}
	w.Header().Set("Content-Type", "application/manifest+json")
	w.Header().Set("Cache-Control", "no-cache")
	_, _ = w.Write(data)
}

// handleServiceWorker 返回 PWA Service Worker 脚本 (严格禁用浏览器 HTTP 缓存，防止死锁旧版本)
func (s *Server) handleServiceWorker(w http.ResponseWriter, r *http.Request) {
	data, err := web.GetStaticFile("sw.js")
	if err != nil {
		http.NotFound(w, r)
		return
	}
	w.Header().Set("Content-Type", "application/javascript; charset=utf-8")
	w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
	w.Header().Set("Pragma", "no-cache")
	w.Header().Set("Expires", "0")
	_, _ = w.Write(data)
}

// handleFavicon 返回网站/PWA应用图标
func (s *Server) handleFavicon(w http.ResponseWriter, r *http.Request) {
	data, err := web.GetStaticFile("img/icon.png")
	if err != nil {
		http.NotFound(w, r)
		return
	}
	w.Header().Set("Content-Type", "image/png")
	w.Header().Set("Cache-Control", "public, max-age=86400")
	_, _ = w.Write(data)
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
	if !s.requireAdminAuth(w, r) {
		return
	}

	newPwd := s.cfg.RegeneratePassword()
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{
		"ok":       true,
		"password": newPwd,
	})
}

// resolveTargetProject 解析请求的目标项目：
// 若传入具体项目ID则定位该项目；若为空或 "default"，则默认返回最近操作的一个项目会话
func (s *Server) resolveTargetProject(requestedID string) config.ProjectConfig {
	if requestedID != "" && requestedID != "default" && requestedID != "latest" {
		if p, ok := s.cfg.GetProject(requestedID); ok {
			return p
		}
	}

	// 检查会话管理器中最近活跃的会话是否对应某个已配置项目
	lastID := session.GetManager().LastActiveID()
	if lastID != "" && lastID != "default" {
		if p, ok := s.cfg.GetProject(lastID); ok {
			return p
		}
	}

	return s.cfg.GetActiveProject()
}

func (s *Server) handleAPIStatus(w http.ResponseWriter, r *http.Request) {
	cfg := s.cfg.Get()
	networks := network.GetAvailableIPs(cfg.Port)

	mgr := session.GetManager()
	reqSessID := r.URL.Query().Get("session")
	activeProj := s.resolveTargetProject(reqSessID)

	// 确保当前默认/活动项目会话已准备就绪
	initCmd := s.cfg.GetAgentCommand(activeProj.Agent, activeProj.AutoApprove)
	defaultSess, _ := mgr.GetOrCreateForDir(activeProj.ID, activeProj.Name, cfg.Shell, activeProj.WorkDir, activeProj.Agent, initCmd, cfg.Language)
	if defaultSess != nil {
		defaultSess.AutoApprove = activeProj.AutoApprove
	}

	var sessInfo map[string]any
	if defaultSess != nil {
		sessInfo = map[string]any{
			"id":            defaultSess.ID,
			"name":          activeProj.Name,
			"pid":           defaultSess.Pid,
			"work_dir":      defaultSess.WorkDir,
			"active_agent":  defaultSess.ActiveAgent,
			"auto_approve":  activeProj.AutoApprove,
			"is_running":    defaultSess.IsRunning,
			"clients_count": len(defaultSess.ListClients()),
		}
	}

	// 组装所有配置项目的会话运行状态列表
	var projectsStatus []map[string]any
	for _, p := range cfg.Projects {
		ps := map[string]any{
			"id":            p.ID,
			"name":          p.Name,
			"work_dir":      p.WorkDir,
			"agent":         p.Agent,
			"auto_approve":  p.AutoApprove,
			"is_active":     p.ID == activeProj.ID,
			"is_running":    false,
			"pid":           0,
			"clients_count": 0,
		}
		if existingSess := mgr.FindByProjectOrDir(p.ID, p.WorkDir); existingSess != nil && existingSess.IsRunning {
			ps["is_running"] = true
			ps["pid"] = existingSess.Pid
			ps["active_agent"] = existingSess.ActiveAgent
			ps["clients_count"] = existingSess.ClientsCount()
			ps["last_active_at"] = existingSess.LastActiveAt
		}
		projectsStatus = append(projectsStatus, ps)
	}

	// 敏感信息保护：非本机且未经鉴权的访问，隐藏密码与鉴权 Token
	isAuthed := isLocalRequest(r) || s.isAuthenticated(r)
	if !isAuthed {
		cfg.Password = "******"
		cfg.AuthToken = ""
	}

	res := map[string]any{
		"version":           config.AppVersion,
		"github_url":        config.GitHubRepoURL,
		"releases_url":      config.GitHubReleases,
		"config":            cfg,
		"networks":          networks,
		"auth_token":        cfg.AuthToken,
		"active_project_id": activeProj.ID,
		"session":           sessInfo,
		"projects_status":   projectsStatus,
		"uptime":            time.Since(s.startTime).String(),
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
	if !s.requireAdminAuth(w, r) {
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
		RemoteAccess  bool   `json:"remote_access"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	oldCfg := s.cfg.Get()
	agentChanged := (req.SelectedAgent != "" && req.SelectedAgent != oldCfg.SelectedAgent)
	workDirChanged := (req.WorkDir != "" && req.WorkDir != oldCfg.WorkDir)
	autoApproveChanged := (req.AutoApprove != oldCfg.AutoApprove && req.WorkDir != "")

	targetLang := req.Language
	if targetLang == "" {
		targetLang = oldCfg.Language
	}

	err := s.cfg.Update(req.Port, req.Password, req.Shell, req.WorkDir, req.SelectedAgent, req.Theme, targetLang, req.AutoApprove, req.KeepAlive, req.RemoteAccess, nil)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{"ok": false, "message": err.Error()})
		return
	}

	restarted := false
	if agentChanged || workDirChanged || autoApproveChanged {
		activeProj := s.cfg.GetActiveProject()
		targetShell := req.Shell
		if targetShell == "" {
			targetShell = oldCfg.Shell
		}
		initCmd := s.cfg.GetAgentCommand(activeProj.Agent, activeProj.AutoApprove)
		mgr := session.GetManager()
		sess, _ := mgr.Restart(activeProj.ID, activeProj.Name, targetShell, activeProj.WorkDir, activeProj.Agent, initCmd, targetLang)
		if sess != nil {
			sess.AutoApprove = activeProj.AutoApprove
		}
		restarted = true
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{
		"ok":        true,
		"restarted": restarted,
		"agent":     req.SelectedAgent,
	})
}

// handleAPIProjectSave 新增或修改项目配置（强制限制：不能同一个目录属于不同项目）
func (s *Server) handleAPIProjectSave(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	if !s.requireAdminAuth(w, r) {
		return
	}

	var req struct {
		ID          string `json:"id"`
		Name        string `json:"name"`
		WorkDir     string `json:"work_dir"`
		Agent       string `json:"agent"`
		AutoApprove bool   `json:"auto_approve"`
		SetActive   bool   `json:"set_active"`
		RestartNow  bool   `json:"restart_now"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// 检查修改前该项目的旧配置，判断其运行中会话是否需要重启
	var oldProj config.ProjectConfig
	hadOld := false
	if req.ID != "" {
		oldProj, hadOld = s.cfg.GetProject(req.ID)
	}

	savedProj, err := s.cfg.UpsertProject(config.ProjectConfig{
		ID:          req.ID,
		Name:        req.Name,
		WorkDir:     req.WorkDir,
		Agent:       req.Agent,
		AutoApprove: req.AutoApprove,
	}, req.SetActive)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"ok":      false,
			"message": err.Error(),
		})
		return
	}

	cfg := s.cfg.Get()
	mgr := session.GetManager()
	existingSess := mgr.FindByProjectOrDir(savedProj.ID, oldProj.WorkDir)
	restarted := false

	if req.RestartNow || (hadOld && existingSess != nil && existingSess.IsRunning &&
		(config.NormalizeWorkDir(oldProj.WorkDir) != config.NormalizeWorkDir(savedProj.WorkDir) ||
			oldProj.Agent != savedProj.Agent ||
			oldProj.AutoApprove != savedProj.AutoApprove)) {
		// 如果该项目配置发生变化且正在运行（或显式要求启动/重启），仅重启该项目会话，不影响后台其他项目会话
		if hadOld && config.NormalizeWorkDir(oldProj.WorkDir) != config.NormalizeWorkDir(savedProj.WorkDir) {
			mgr.CloseSession(savedProj.ID, oldProj.WorkDir)
		}
		initCmd := s.cfg.GetAgentCommand(savedProj.Agent, savedProj.AutoApprove)
		sess, err := mgr.Restart(savedProj.ID, savedProj.Name, cfg.Shell, savedProj.WorkDir, savedProj.Agent, initCmd, cfg.Language)
		if err == nil && sess != nil {
			sess.AutoApprove = savedProj.AutoApprove
			restarted = true
		}
	} else if existingSess != nil {
		existingSess.Name = savedProj.Name
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{
		"ok":        true,
		"project":   savedProj,
		"restarted": restarted,
	})
}

// handleAPIProjectDelete 删除项目配置并清理其对应会话
func (s *Server) handleAPIProjectDelete(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	if !s.requireAdminAuth(w, r) {
		return
	}

	var req struct {
		ID string `json:"id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	proj, ok := s.cfg.GetProject(req.ID)
	if !ok {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{"ok": false, "message": "未找到该项目"})
		return
	}

	if err := s.cfg.DeleteProject(req.ID); err != nil {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{"ok": false, "message": err.Error()})
		return
	}

	session.GetManager().CloseSession(proj.ID, proj.WorkDir)

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{
		"ok":                true,
		"active_project_id": s.cfg.GetActiveProject().ID,
	})
}

// handleAPIProjectSelect 切换当前活跃项目会话（支持从终端或管理面板切换，原会话继续保持在后台运行）
func (s *Server) handleAPIProjectSelect(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	if !isLocalRequest(r) && !s.isAuthenticated(r) {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var req struct {
		ID        string `json:"id"`
		ProjectID string `json:"project_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	targetID := req.ProjectID
	if targetID == "" {
		targetID = req.ID
	}

	proj, ok := s.cfg.SetActiveProject(targetID)
	if !ok {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{"ok": false, "message": "项目不存在"})
		return
	}

	cfg := s.cfg.Get()
	initCmd := s.cfg.GetAgentCommand(proj.Agent, proj.AutoApprove)
	sess, err := session.GetManager().GetOrCreateForDir(proj.ID, proj.Name, cfg.Shell, proj.WorkDir, proj.Agent, initCmd, cfg.Language)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{"ok": false, "message": err.Error()})
		return
	}
	sess.AutoApprove = proj.AutoApprove
	sess.Touch()

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{
		"ok":      true,
		"project": proj,
		"session": map[string]any{
			"id":            sess.ID,
			"name":          proj.Name,
			"pid":           sess.Pid,
			"work_dir":      sess.WorkDir,
			"active_agent":  sess.ActiveAgent,
			"auto_approve":  proj.AutoApprove,
			"is_running":    sess.IsRunning,
			"clients_count": sess.ClientsCount(),
		},
	})
}

func (s *Server) handleAPISessionRestart(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	if !isLocalRequest(r) && !s.isAuthenticated(r) {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var req struct {
		ProjectID   string `json:"project_id"`
		Agent       string `json:"agent"`
		WorkDir     string `json:"work_dir"`
		AutoApprove *bool  `json:"auto_approve"`
	}
	_ = json.NewDecoder(r.Body).Decode(&req)

	targetProj := s.resolveTargetProject(req.ProjectID)
	if req.WorkDir != "" {
		targetProj.WorkDir = req.WorkDir
	}
	if req.Agent != "" {
		targetProj.Agent = req.Agent
	}
	if req.AutoApprove != nil {
		targetProj.AutoApprove = *req.AutoApprove
	}

	savedProj, err := s.cfg.UpsertProject(targetProj, true)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{"ok": false, "message": err.Error()})
		return
	}

	cfg := s.cfg.Get()
	initCmd := s.cfg.GetAgentCommand(savedProj.Agent, savedProj.AutoApprove)
	mgr := session.GetManager()
	sess, err := mgr.Restart(savedProj.ID, savedProj.Name, cfg.Shell, savedProj.WorkDir, savedProj.Agent, initCmd, cfg.Language)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{"ok": false, "message": err.Error()})
		return
	}
	if sess != nil {
		sess.AutoApprove = savedProj.AutoApprove
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{
		"ok":      true,
		"project": savedProj,
	})
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

	reqSessID := r.URL.Query().Get("session")
	targetProj := s.resolveTargetProject(reqSessID)
	// 记录当前连接的项目为最近操作的会话
	s.cfg.SetActiveProject(targetProj.ID)

	cfg := s.cfg.Get()
	initCmd := s.cfg.GetAgentCommand(targetProj.Agent, targetProj.AutoApprove)
	sess, err := session.GetManager().GetOrCreateForDir(targetProj.ID, targetProj.Name, cfg.Shell, targetProj.WorkDir, targetProj.Agent, initCmd, cfg.Language)
	if err != nil {
		_ = conn.WriteMessage(websocket.TextMessage, []byte("启动终端失败: "+err.Error()))
		_ = conn.Close()
		return
	}
	sess.AutoApprove = targetProj.AutoApprove

	// 挂载连接到会话
	sess.Attach(conn)
}

// handleAPIAppInstall 一键在当前主机生成/安装桌面 GUI 应用程序 (macOS .app / Windows .lnk / Linux .desktop)
func (s *Server) handleAPIAppInstall(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	if !s.requireAdminAuth(w, r) {
		return
	}

	installedPath, err := appbundle.InstallDesktopApp()
	w.Header().Set("Content-Type", "application/json")
	if err != nil {
		_ = json.NewEncoder(w).Encode(map[string]any{
			"ok":      false,
			"message": err.Error(),
		})
		return
	}

	_ = json.NewEncoder(w).Encode(map[string]any{
		"ok":   true,
		"path": installedPath,
	})
}

