package config

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"os"
	"path/filepath"
	"sync"
)

const (
	AppVersion     = "v1.0.3"
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

type ConfigData struct {
	Port           int           `json:"port"`
	Password       string        `json:"password"`
	AuthToken      string        `json:"auth_token"`
	Shell          string        `json:"shell"`
	WorkDir        string        `json:"work_dir"`
	SelectedAgent  string        `json:"selected_agent"`
	AutoApprove    bool          `json:"auto_approve"` // 自动确认(关闭沙箱)，默认false
	RemoteAccess   bool          `json:"remote_access"` // 远程管理访问(开启后允许外部IP访问管理面板，需密码鉴权)，默认false
	Theme          string        `json:"theme"`          // 终端配色方案
	Language       string        `json:"language"`       // 界面语言: "en" (默认) 或 "zh"
	KeepAlive      bool          `json:"keep_alive"`
	AutoOpen       bool          `json:"auto_open"`
	CustomCommands []CommandItem `json:"custom_commands"`
	Agents         []AgentPreset `json:"agents"`
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

	return &Config{
		ConfigData: ConfigData{
			Port:          20996,
			Password:      generateRandomPassword(8), // 默认强制生成 8 位随机密码
			AuthToken:     generateRandomToken(16),
			Shell:         shell,
			WorkDir:       cwd,
			SelectedAgent: "claude", // 默认选用常用 claude
			AutoApprove:   false,
			RemoteAccess:  false, // 默认仅限本机访问
			Theme:         "github-dark", // 默认配色方案
			Language:      "en",            // 默认语言为英文
			KeepAlive:     true,
			AutoOpen:      true,
			Agents:        DefaultAgents(),
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

	return cfg, nil
}

func (c *Config) Save() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	data, err := json.MarshalIndent(c.ConfigData, "", "  ")
	if err != nil {
		return err
	}

	if c.filePath == "" {
		c.filePath = getConfigPath()
	}

	dir := filepath.Dir(c.filePath)
	if dir != "" && dir != "." {
		_ = os.MkdirAll(dir, 0755)
	}

	return os.WriteFile(c.filePath, data, 0644)
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
	snapshot := c.ConfigData
	c.mu.Unlock()

	err := c.Save()
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
	return c.ConfigData
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
