package config

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

const (
	AppVersion     = "v2.1.0"
	GitHubRepoURL  = "https://github.com/tim-today/tim-agent"
	GitHubReleases = "https://github.com/tim-today/tim-agent/releases"
	GitHubAPIURL   = "https://api.github.com/repos/tim-today/tim-agent/releases/latest"
)

type CommandItem struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Command   string `json:"command"`
	AutoEnter bool   `json:"auto_enter"`
}

type AgentPreset struct {
	ID              string   `json:"id"`
	Name            string   `json:"name"`
	Command         string   `json:"command"`
	AutoApproveFlag string   `json:"auto_approve_flag"` // 关闭沙箱/自动确认的附加参数
	Description     string   `json:"description"`
	Shortcuts       []string `json:"shortcuts"`
}

// DeployEnvConfig 单个部署环境配置（如：生产环境、测试环境、本地测试）
type DeployEnvConfig struct {
	ID            string `json:"id"`
	EnvName       string `json:"env_name"`        // 环境名称: 如 "生产环境", "测试环境", "本地测试"
	Method        string `json:"method"`          // 部署方式: "ssh" (SSH/SCP), "sftp_ftp" (FTP/SFTP), "local_cmd" (仅本地命令)
	Host          string `json:"host"`            // 服务器 IP 或域名
	Port          int    `json:"port"`            // 端口 (默认 SSH 22 / FTP 21)
	Username      string `json:"username"`        // 服务器登录用户名
	AuthType      string `json:"auth_type"`       // 认证方式: "password" (密码/Token), "ssh_key" (密钥)
	Password      string `json:"password"`        // 服务器密码或上传 Token
	SSHKeyPath    string `json:"ssh_key_path"`    // SSH 私钥文件路径
	SSHKeyContent string `json:"ssh_key_content"` // 直接粘贴的 SSH 私钥证书文本内容
	RemoteDir     string `json:"remote_dir"`      // 远程服务器目标目录 (如 /var/www/html 或 /data/apk)
	CustomPrompt  string `json:"custom_prompt"`   // 自定义部署提示词 (多行文本)
	IgnoreFiles   string `json:"ignore_files"`    // 忽略文件列表 (多行文本，每行一个)
}

// DeployAppConfig 项目下的单个部署应用配置（如：前端H5、后端API、APK打包）
type DeployAppConfig struct {
	ID   string            `json:"id"`
	Name string            `json:"name"` // 应用名称: 如 "前端H5", "后端API", "APK打包"
	Type string            `json:"type"` // 应用类型: "web_h5", "backend_api", "android_apk", "docker", "custom"
	Envs []DeployEnvConfig `json:"envs"` // 多环境列表
}

// ProjectConfig 单个项目的独立配置（支持多项目多会话切换与多应用多环境部署）
type ProjectConfig struct {
	ID          string            `json:"id"`
	Name        string            `json:"name"`
	WorkDir     string            `json:"work_dir"`
	Agent       string            `json:"agent"`
	AutoApprove bool              `json:"auto_approve"`
	DeployApps  []DeployAppConfig `json:"deploy_apps"`
}

type ConfigData struct {
	Port            int             `json:"port"`
	Password        string          `json:"password"`
	AuthToken       string          `json:"auth_token"`
	Shell           string          `json:"shell"`
	WorkDir         string          `json:"work_dir"`
	SelectedAgent   string          `json:"selected_agent"`
	AutoApprove     bool            `json:"auto_approve"`      // 自动确认(关闭沙箱)，默认false
	RemoteAccess    bool            `json:"remote_access"`     // 远程管理访问(开启后允许外部IP访问管理面板，需密码鉴权)，默认false
	Theme           string          `json:"theme"`             // 终端配色方案
	Language        string          `json:"language"`          // 界面语言: "en" (默认) 或 "zh"
	KeepAlive       bool            `json:"keep_alive"`
	AutoOpen        bool            `json:"auto_open"`
	ActiveProjectID string          `json:"active_project_id"` // 当前/最近操作的项目会话ID
	Projects        []ProjectConfig `json:"projects"`          // 多项目配置列表
	CustomCommands  []CommandItem   `json:"custom_commands"`
	Agents          []AgentPreset   `json:"agents"`
}

type Config struct {
	ConfigData
	mu        sync.RWMutex
	filePath  string
	listeners []func(ConfigData)
}

// AddListener 注册配置变更监听器
func (c *Config) AddListener(fn func(ConfigData)) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.listeners = append(c.listeners, fn)
}

func (c *Config) notifyListeners(data ConfigData) {
	c.mu.RLock()
	subs := make([]func(ConfigData), len(c.listeners))
	copy(subs, c.listeners)
	c.mu.RUnlock()

	for _, fn := range subs {
		if fn != nil {
			go fn(data)
		}
	}
}

func DefaultAgents() []AgentPreset {
	return []AgentPreset{
		{
			ID:              "claude",
			Name:            "Claude Code",
			Command:         "claude",
			AutoApproveFlag: "--dangerously-skip-permissions",
			Description:     "Anthropic Claude Code CLI 辅助开发",
			Shortcuts:       []string{"/help", "/compact", "/cost", "/review", "exit"},
		},
		{
			ID:              "codex",
			Name:            "Codex CLI",
			Command:         "codex",
			AutoApproveFlag: "--dangerously-bypass-approvals-and-sandbox",
			Description:     "OpenAI Codex 命令行开发工具",
			Shortcuts:       []string{"--help", "generate", "explain", "review"},
		},
		{
			ID:              "opcode",
			Name:            "OpenCode (opcode)",
			Command:         "opcode",
			AutoApproveFlag: "--yes",
			Description:     "OpenCode 终端 AI Agent",
			Shortcuts:       []string{"help", "run", "status"},
		},
		{
			ID:              "pi",
			Name:            "Pi CLI",
			Command:         "pi",
			AutoApproveFlag: "-y",
			Description:     "Inflection Pi 命令行对话与辅助工具",
			Shortcuts:       []string{"--help", "quit"},
		},
		{
			ID:              "gemini",
			Name:            "Gemini CLI",
			Command:         "gemini",
			AutoApproveFlag: "--auto-approve",
			Description:     "Google Gemini 命令行代码助手",
			Shortcuts:       []string{"--help", "chat", "code"},
		},
		{
			ID:              "shell",
			Name:            "普通 Shell",
			Command:         "",
			AutoApproveFlag: "",
			Description:     "系统原生交互式 Shell (无预设 Agent)",
			Shortcuts:       []string{"ls -la", "git status", "git diff", "clear"},
		},
	}
}

// NormalizeWorkDir 规范化目录路径（清理末尾斜杠并解析软链），用于判断目录唯一性
func NormalizeWorkDir(p string) string {
	p = strings.TrimSpace(p)
	if p == "" {
		return ""
	}
	clean := filepath.Clean(p)
	if realPath, err := filepath.EvalSymlinks(clean); err == nil {
		return realPath
	}
	return clean
}

func defaultProjectFrom(workDir, agent string, autoApprove bool) ProjectConfig {
	name := filepath.Base(workDir)
	if name == "" || name == "." || name == "/" || name == "\\" {
		name = "Default Project"
	}
	if agent == "" {
		agent = "claude"
	}
	return ProjectConfig{
		ID:          "proj-default",
		Name:        name,
		WorkDir:     workDir,
		Agent:       agent,
		AutoApprove: autoApprove,
	}
}

func DefaultConfig() *Config {
	shell := os.Getenv("SHELL")
	if shell == "" {
		if _, err := os.Stat("/bin/zsh"); err == nil {
			shell = "/bin/zsh"
		} else if _, err := os.Stat("/bin/bash"); err == nil {
			shell = "/bin/bash"
		} else {
			shell = "/bin/sh"
		}
	}

	cwd, err := os.Getwd()
	if err != nil {
		cwd, _ = os.UserHomeDir()
	}

	defProj := defaultProjectFrom(cwd, "claude", false)

	return &Config{
		ConfigData: ConfigData{
			Port:            20996,
			Password:        generateRandomPassword(8), // 默认强制生成 8 位随机密码
			AuthToken:       generateRandomToken(16),
			Shell:           shell,
			WorkDir:         cwd,
			SelectedAgent:   "claude", // 默认选用常用 claude
			AutoApprove:     false,
			RemoteAccess:    false,         // 默认仅限本机访问
			Theme:           "github-dark", // 默认配色方案
			Language:        "en",          // 默认语言为英文
			KeepAlive:       true,
			AutoOpen:        true,
			ActiveProjectID: defProj.ID,
			Projects:        []ProjectConfig{defProj},
			Agents:          DefaultAgents(),
			CustomCommands: []CommandItem{
				{ID: "cmd-1", Name: "Claude", Command: "claude", AutoEnter: true},
				{ID: "cmd-2", Name: "Codex", Command: "codex", AutoEnter: true},
				{ID: "cmd-3", Name: "Git Status", Command: "git status", AutoEnter: true},
				{ID: "cmd-4", Name: "Git Diff", Command: "git diff", AutoEnter: true},
				{ID: "cmd-5", Name: "Git Log", Command: "git log -n 5 --oneline", AutoEnter: true},
				{ID: "cmd-6", Name: "Clear", Command: "clear", AutoEnter: true},
				{ID: "cmd-7", Name: "List", Command: "ls -la", AutoEnter: true},
			},
		},
	}
}

// generateRandomPassword 生成无歧义字符的高可读性随机密码 (排除 0, O, 1, l, I)
func generateRandomPassword(length int) string {
	const charset = "23456789abcdefghjkmnpqrstuvwxyzABCDEFGHJKLMNPQRSTUVWXYZ"
	b := make([]byte, length)
	randBytes := make([]byte, length)
	if _, err := rand.Read(randBytes); err != nil {
		return "tim88888"
	}
	for i := 0; i < length; i++ {
		b[i] = charset[int(randBytes[i])%len(charset)]
	}
	return string(b)
}

func generateRandomToken(length int) string {
	b := make([]byte, length)
	if _, err := rand.Read(b); err != nil {
		return "tim-agent-secret-key"
	}
	return hex.EncodeToString(b)
}

func getConfigPath() string {
	localPath := "tim-agent.json"
	// 优先使用当前目录下的 tim-agent.json
	if _, err := os.Stat(localPath); err == nil {
		return localPath
	}
	// 尝试在当前目录创建
	if f, err := os.OpenFile(localPath, os.O_CREATE|os.O_RDWR, 0644); err == nil {
		_ = f.Close()
		return localPath
	}

	home, err := os.UserHomeDir()
	if err == nil {
		dir := filepath.Join(home, ".tim-agent")
		if err := os.MkdirAll(dir, 0755); err == nil {
			return filepath.Join(dir, "config.json")
		}
	}
	return localPath
}

func LoadConfig() (*Config, error) {
	cfgPath := getConfigPath()
	cfg := DefaultConfig()
	cfg.filePath = cfgPath

	data, err := os.ReadFile(cfgPath)
	if err != nil {
		if os.IsNotExist(err) {
			_ = cfg.Save()
			return cfg, nil
		}
		return cfg, err
	}

	if len(data) == 0 {
		_ = cfg.Save()
		return cfg, nil
	}

	if err := json.Unmarshal(data, &cfg.ConfigData); err != nil {
		return cfg, err
	}

	if cfg.Password == "" {
		cfg.Password = generateRandomPassword(8)
		_ = cfg.Save()
	}

	if cfg.AuthToken == "" {
		cfg.AuthToken = generateRandomToken(16)
		_ = cfg.Save()
	}

	if cfg.WorkDir == "" {
		cfg.WorkDir, _ = os.Getwd()
		_ = cfg.Save()
	}

	if len(cfg.Agents) == 0 {
		cfg.Agents = DefaultAgents()
		_ = cfg.Save()
	} else {
		// 自动平滑升级旧版 Agent 参数（例如 codex --full-auto 现已过时）
		updated := false
		for i := range cfg.Agents {
			if cfg.Agents[i].ID == "codex" && (cfg.Agents[i].AutoApproveFlag == "--full-auto" || cfg.Agents[i].AutoApproveFlag == "") {
				cfg.Agents[i].AutoApproveFlag = "--dangerously-bypass-approvals-and-sandbox"
				updated = true
			}
		}
		if updated {
			_ = cfg.Save()
		}
	}

	if cfg.Theme == "" {
		cfg.Theme = "github-dark"
		_ = cfg.Save()
	}

	if cfg.Language == "" {
		cfg.Language = "en"
		_ = cfg.Save()
	}

	if cfg.Port == 0 || cfg.Port == 8080 {
		cfg.Port = 20996
		_ = cfg.Save()
	}

	// 初始化或修复多项目列表
	needSaveProjects := false
	if len(cfg.Projects) == 0 {
		defProj := defaultProjectFrom(cfg.WorkDir, cfg.SelectedAgent, cfg.AutoApprove)
		cfg.Projects = []ProjectConfig{defProj}
		cfg.ActiveProjectID = defProj.ID
		needSaveProjects = true
	} else {
		// 去重同一目录的项目并补全缺失字段
		seenDirs := make(map[string]bool)
		var validProjs []ProjectConfig
		for _, p := range cfg.Projects {
			p.WorkDir = strings.TrimSpace(p.WorkDir)
			if p.WorkDir == "" {
				continue
			}
			norm := NormalizeWorkDir(p.WorkDir)
			if seenDirs[norm] {
				needSaveProjects = true
				continue
			}
			seenDirs[norm] = true
			if p.ID == "" {
				p.ID = "proj-" + generateRandomToken(4)
				needSaveProjects = true
			}
			if p.Name == "" {
				p.Name = filepath.Base(p.WorkDir)
				needSaveProjects = true
			}
			if p.Agent == "" {
				p.Agent = "claude"
				needSaveProjects = true
			}
			normalizeProjectSubConfigs(&p)
			_ = SyncProjectVault(p)
			validProjs = append(validProjs, p)
		}
		if len(validProjs) == 0 {
			defProj := defaultProjectFrom(cfg.WorkDir, cfg.SelectedAgent, cfg.AutoApprove)
			validProjs = []ProjectConfig{defProj}
			needSaveProjects = true
		}
		cfg.Projects = validProjs
	}

	// 确保 ActiveProjectID 有效并同步顶层兼容字段
	foundActive := false
	for _, p := range cfg.Projects {
		if p.ID == cfg.ActiveProjectID {
			foundActive = true
			cfg.WorkDir = p.WorkDir
			cfg.SelectedAgent = p.Agent
			cfg.AutoApprove = p.AutoApprove
			break
		}
	}
	if !foundActive && len(cfg.Projects) > 0 {
		cfg.ActiveProjectID = cfg.Projects[0].ID
		cfg.WorkDir = cfg.Projects[0].WorkDir
		cfg.SelectedAgent = cfg.Projects[0].Agent
		cfg.AutoApprove = cfg.Projects[0].AutoApprove
		needSaveProjects = true
	}

	if needSaveProjects {
		_ = cfg.Save()
	}

	return cfg, nil
}

func (c *Config) Save() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	return c.saveLocked()
}

func (c *Config) saveLocked() error {
	data, err := json.MarshalIndent(c.ConfigData, "", "  ")
	if err != nil {
		return err
	}

	if c.filePath == "" {
		return nil
	}

	dir := filepath.Dir(c.filePath)
	if dir != "" && dir != "." {
		_ = os.MkdirAll(dir, 0755)
	}

	return os.WriteFile(c.filePath, data, 0644)
}

func (c *Config) syncTopFieldsFromActiveLocked() {
	if len(c.Projects) == 0 {
		defProj := defaultProjectFrom(c.WorkDir, c.SelectedAgent, c.AutoApprove)
		c.Projects = []ProjectConfig{defProj}
		c.ActiveProjectID = defProj.ID
	}
	for _, p := range c.Projects {
		if p.ID == c.ActiveProjectID {
			c.WorkDir = p.WorkDir
			c.SelectedAgent = p.Agent
			c.AutoApprove = p.AutoApprove
			return
		}
	}
	c.ActiveProjectID = c.Projects[0].ID
	c.WorkDir = c.Projects[0].WorkDir
	c.SelectedAgent = c.Projects[0].Agent
	c.AutoApprove = c.Projects[0].AutoApprove
}

func (c *Config) Update(port int, password, shell, workDir, selectedAgent, theme, language string, autoApprove, keepAlive, remoteAccess bool, commands []CommandItem) error {
	c.mu.Lock()
	if port > 0 && port < 65536 {
		c.Port = port
	}
	if password != "" {
		c.Password = password
	} else if c.Password == "" {
		c.Password = generateRandomPassword(8)
	}
	if shell != "" {
		c.Shell = shell
	}
	if workDir != "" {
		c.WorkDir = workDir
	}
	if selectedAgent != "" {
		c.SelectedAgent = selectedAgent
	}
	if theme != "" {
		c.Theme = theme
	}
	if language != "" {
		c.Language = language
	}
	c.AutoApprove = autoApprove
	c.KeepAlive = keepAlive
	c.RemoteAccess = remoteAccess
	if commands != nil {
		c.CustomCommands = commands
	}

	// 若传入了 workDir 或 selectedAgent，同步更新当前活动项目（同时校验不与其它项目目录冲突）
	if len(c.Projects) == 0 {
		defProj := defaultProjectFrom(c.WorkDir, c.SelectedAgent, c.AutoApprove)
		c.Projects = []ProjectConfig{defProj}
		c.ActiveProjectID = defProj.ID
	} else if workDir != "" || selectedAgent != "" {
		activeIdx := 0
		for i, p := range c.Projects {
			if p.ID == c.ActiveProjectID {
				activeIdx = i
				break
			}
		}
		if workDir != "" {
			normTarget := NormalizeWorkDir(workDir)
			conflict := false
			for i, p := range c.Projects {
				if i != activeIdx && NormalizeWorkDir(p.WorkDir) == normTarget {
					conflict = true
					break
				}
			}
			if !conflict {
				c.Projects[activeIdx].WorkDir = workDir
			}
		}
		if selectedAgent != "" {
			c.Projects[activeIdx].Agent = selectedAgent
		}
		c.Projects[activeIdx].AutoApprove = autoApprove
	}
	c.syncTopFieldsFromActiveLocked()

	err := c.saveLocked()
	snapshot := c.ConfigData
	c.mu.Unlock()

	if err == nil {
		c.notifyListeners(snapshot)
	}
	return err
}

// GetActiveProject 返回当前活动（最近操作）的项目配置
func (c *Config) GetActiveProject() ProjectConfig {
	c.mu.RLock()
	defer c.mu.RUnlock()

	for _, p := range c.Projects {
		if p.ID == c.ActiveProjectID {
			return p
		}
	}
	if len(c.Projects) > 0 {
		return c.Projects[0]
	}
	return defaultProjectFrom(c.WorkDir, c.SelectedAgent, c.AutoApprove)
}

// GetProject 根据 ID 或目录查找项目
func (c *Config) GetProject(idOrDir string) (ProjectConfig, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	for _, p := range c.Projects {
		if p.ID == idOrDir {
			return p, true
		}
	}
	if idOrDir != "" && idOrDir != "default" {
		norm := NormalizeWorkDir(idOrDir)
		for _, p := range c.Projects {
			if NormalizeWorkDir(p.WorkDir) == norm {
				return p, true
			}
		}
	}
	return ProjectConfig{}, false
}

// SetActiveProject 设置当前/最近操作的项目会话
func (c *Config) SetActiveProject(id string) (ProjectConfig, bool) {
	c.mu.Lock()
	var target ProjectConfig
	found := false
	for _, p := range c.Projects {
		if p.ID == id {
			target = p
			found = true
			break
		}
	}
	if !found {
		c.mu.Unlock()
		return ProjectConfig{}, false
	}

	changed := c.ActiveProjectID != id
	c.ActiveProjectID = id
	c.syncTopFieldsFromActiveLocked()
	if changed {
		_ = c.saveLocked()
	}
	snapshot := c.ConfigData
	c.mu.Unlock()

	if changed {
		c.notifyListeners(snapshot)
	}
	return target, true
}

// UpsertProject 新增或修改单个项目配置（强制校验：不能同一个目录属于不同项目）
func (c *Config) UpsertProject(proj ProjectConfig, setActive bool) (ProjectConfig, error) {
	c.mu.Lock()
	proj.Name = strings.TrimSpace(proj.Name)
	proj.WorkDir = strings.TrimSpace(proj.WorkDir)
	if proj.WorkDir == "" {
		c.mu.Unlock()
		return ProjectConfig{}, fmt.Errorf("项目工作目录 (WorkDir) 不能为空")
	}
	if proj.Name == "" {
		proj.Name = filepath.Base(proj.WorkDir)
		if proj.Name == "" || proj.Name == "." || proj.Name == "/" {
			proj.Name = "未命名项目"
		}
	}
	if proj.Agent == "" {
		proj.Agent = "claude"
	}
	normalizeProjectSubConfigs(&proj)

	normDir := NormalizeWorkDir(proj.WorkDir)

	// 校验目录唯一性：不能同一个目录属于不同项目
	for _, existing := range c.Projects {
		if existing.ID != proj.ID && NormalizeWorkDir(existing.WorkDir) == normDir {
			c.mu.Unlock()
			return ProjectConfig{}, fmt.Errorf("工作目录 [%s] 已属于项目「%s」，不能将同一个目录分配给不同项目", proj.WorkDir, existing.Name)
		}
	}

	updated := false
	if proj.ID != "" {
		for i := range c.Projects {
			if c.Projects[i].ID == proj.ID {
				// 若调用方仅修改了基础字段（如终端切换 Agent 时未传 DeployApps），自动保留该项目已有的 DeployApps
				if proj.DeployApps == nil && len(c.Projects[i].DeployApps) > 0 {
					proj.DeployApps = c.Projects[i].DeployApps
				}
				c.Projects[i] = proj
				updated = true
				break
			}
		}
	}
	if !updated {
		if proj.ID == "" {
			proj.ID = "proj-" + generateRandomToken(4)
		}
		c.Projects = append(c.Projects, proj)
	}

	_ = SyncProjectVault(proj)

	if setActive || c.ActiveProjectID == "" || c.ActiveProjectID == proj.ID {
		if setActive || c.ActiveProjectID == "" {
			c.ActiveProjectID = proj.ID
		}
	}
	c.syncTopFieldsFromActiveLocked()
	err := c.saveLocked()
	snapshot := c.ConfigData
	c.mu.Unlock()

	if err == nil {
		c.notifyListeners(snapshot)
	}
	return proj, err
}

// DeleteProject 删除指定项目（至少保留一个项目）
func (c *Config) DeleteProject(id string) error {
	c.mu.Lock()
	if len(c.Projects) <= 1 {
		c.mu.Unlock()
		return fmt.Errorf("至少需要保留一个项目配置")
	}

	idx := -1
	for i, p := range c.Projects {
		if p.ID == id {
			idx = i
			break
		}
	}
	if idx == -1 {
		c.mu.Unlock()
		return fmt.Errorf("未找到该项目")
	}

	c.Projects = append(c.Projects[:idx], c.Projects[idx+1:]...)
	if c.ActiveProjectID == id {
		c.ActiveProjectID = c.Projects[0].ID
	}
	c.syncTopFieldsFromActiveLocked()
	err := c.saveLocked()
	snapshot := c.ConfigData
	c.mu.Unlock()

	if err == nil {
		c.notifyListeners(snapshot)
	}
	return err
}

func (c *Config) RegeneratePassword() string {
	c.mu.Lock()
	c.Password = generateRandomPassword(8)
	snapshot := c.ConfigData
	c.mu.Unlock()
	_ = c.Save()
	c.notifyListeners(snapshot)
	return c.Password
}

func (c *Config) Get() ConfigData {
	c.mu.RLock()
	defer c.mu.RUnlock()
	out := c.ConfigData
	if len(c.Projects) > 0 {
		out.Projects = make([]ProjectConfig, len(c.Projects))
		for i, p := range c.Projects {
			pCopy := p
			if len(p.DeployApps) > 0 {
				pCopy.DeployApps = make([]DeployAppConfig, len(p.DeployApps))
				for j, app := range p.DeployApps {
					appCopy := app
					if len(app.Envs) > 0 {
						appCopy.Envs = make([]DeployEnvConfig, len(app.Envs))
						copy(appCopy.Envs, app.Envs)
					}
					pCopy.DeployApps[j] = appCopy
				}
			}
			out.Projects[i] = pCopy
		}
	}
	return out
}

// GetAgentCommand 获取当前选定 Agent 对应的启动命令（若开启 autoApprove 则追加对应免确认/关沙箱参数）
func (c *Config) GetAgentCommand(agentID string, autoApprove bool) string {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if agentID == "" {
		agentID = c.SelectedAgent
	}

	for _, a := range c.Agents {
		if a.ID == agentID {
			cmd := a.Command
			if cmd != "" && autoApprove && a.AutoApproveFlag != "" {
				cmd = cmd + " " + a.AutoApproveFlag
			}
			return cmd
		}
	}
	return ""
}

// normalizeProjectSubConfigs 规范化项目下的多应用多环境部署配置
func normalizeProjectSubConfigs(proj *ProjectConfig) {
	for i := range proj.DeployApps {
		app := &proj.DeployApps[i]
		app.Name = strings.TrimSpace(app.Name)
		if app.Name == "" {
			app.Name = fmt.Sprintf("应用-%d", i+1)
		}
		if app.ID == "" {
			app.ID = fmt.Sprintf("app-%d-%s", i+1, generateRandomToken(3))
		}
		if app.Type == "" {
			app.Type = "custom"
		}
		for j := range app.Envs {
			env := &app.Envs[j]
			env.EnvName = strings.TrimSpace(env.EnvName)
			if env.EnvName == "" {
				env.EnvName = fmt.Sprintf("环境-%d", j+1)
			}
			if env.ID == "" {
				env.ID = fmt.Sprintf("env-%d-%s", j+1, generateRandomToken(3))
			}
			if env.Method == "" {
				env.Method = "ssh"
			}
			if env.Port <= 0 {
				if env.Method == "sftp_ftp" {
					env.Port = 21
				} else {
					env.Port = 22
				}
			}
			if env.AuthType == "" {
				env.AuthType = "password"
			}
		}
	}
}

