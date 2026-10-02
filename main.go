package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/tim-today/tim-agent/pkg/config"
	"github.com/tim-today/tim-agent/pkg/daemon"
	"github.com/tim-today/tim-agent/pkg/network"
	"github.com/tim-today/tim-agent/pkg/server"
	"github.com/tim-today/tim-agent/pkg/session"
	"github.com/tim-today/tim-agent/pkg/tray"
)

const Version = config.AppVersion

func printBanner(port int, token, workDir, agent, password string) {
	fmt.Println("================================================================")
	fmt.Printf("   🚀 Tim-Agent 在线终端服务 (%s)\n", Version)
	fmt.Println("   兼容 macOS (Apple Silicon & Intel) / Windows / Linux")
	fmt.Printf("   📁 绑定目录: %s\n", workDir)
	fmt.Printf("   🤖 代理软件: %s\n", agent)
	fmt.Printf("   🔑 访问密码: %s (浏览器Web访问输入一次自动保存)\n", password)
	fmt.Println("================================================================")

	ips := network.GetAvailableIPs(port)
	fmt.Println("  [可用访问地址]:")
	for _, ip := range ips {
		tag := "LAN"
		if ip.IsVPN {
			tag = "Tailscale"
		} else if ip.Interface == "lo0" || ip.IP == "127.0.0.1" {
			tag = "Local"
		}
		url := fmt.Sprintf("%s/terminal?token=%s", ip.URL, token)
		fmt.Printf("   • [%-9s] %s\n", tag, url)
	}

	fmt.Println("----------------------------------------------------------------")
	fmt.Printf("  控制面板地址: http://127.0.0.1:%d/\n", port)
	fmt.Println("  已常驻托盘监听与守护自愈，关闭终端窗口或断网保持永不中断")
	fmt.Println("================================================================")
}

func main() {
	// 无论以何种方式运行，首先免疫终端挂断与管道中断信号 (SIGHUP / SIGPIPE)
	// 确保在 macOS / Linux 下双击打开后关掉 Terminal 窗口时，服务依然在后台稳固运行，绝不闪退
	daemon.IgnoreHangupSignals()

	flagPort := flag.Int("port", 0, "监听端口 (默认使用配置文件或 8080)")
	flagPassword := flag.String("password", "", "访问密码 (可选)")
	flagShell := flag.String("shell", "", "Shell路径 (如 /bin/zsh, powershell.exe)")
	flagDir := flag.String("dir", "", "绑定工作目录 (默认当前目录)")
	flagAgent := flag.String("agent", "", "默认代理命令 (如 claude, codex, opcode, pi, gemini)")
	flagAutoApprove := flag.Bool("auto-approve", false, "自动确认/关闭沙箱 (Claude/Codex免手动交互确认)")
	flagNoOpen := flag.Bool("no-open", false, "启动后不自动打开浏览器")
	flagDaemon := flag.Bool("daemon", false, "以独立 Supervisor 守护进程模式运行 (主进程自动监控与自愈重启)")
	flagVersion := flag.Bool("version", false, "查看版本号")
	flag.Parse()

	if *flagVersion {
		fmt.Printf("tim-agent version %s\n", Version)
		return
	}

	// 若显式指定了 -daemon 守护模式，且当前非被监控子进程，则启动守护监控器
	if *flagDaemon && !daemon.IsWorkerProcess() {
		daemon.RunSupervisor()
		return
	}

	// 1. 加载配置
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("加载配置失败: %v", err)
	}

	if *flagPort > 0 {
		cfg.Port = *flagPort
	}
	if *flagPassword != "" {
		cfg.Password = *flagPassword
	}
	if *flagShell != "" {
		cfg.Shell = *flagShell
	}
	if *flagDir != "" {
		cfg.WorkDir = *flagDir
	}
	if *flagAgent != "" {
		cfg.SelectedAgent = *flagAgent
	}
	if *flagAutoApprove {
		cfg.AutoApprove = true
	}

	currentCfg := cfg.Get()
	initCmd := cfg.GetAgentCommand(currentCfg.SelectedAgent, currentCfg.AutoApprove)

	// 2. 打印控制台横幅
	printBanner(currentCfg.Port, currentCfg.AuthToken, currentCfg.WorkDir, currentCfg.SelectedAgent, currentCfg.Password)

	// 3. 预初始化主会话并绑定目录与拉起 Agent
	mgr := session.GetManager()
	_, err = mgr.GetOrCreate("default", "主会话", currentCfg.Shell, currentCfg.WorkDir, currentCfg.SelectedAgent, initCmd, currentCfg.Language)
	if err != nil {
		log.Printf("警告: 预创建终端会话失败: %v", err)
	}

	// 4. 启动 HTTP & WebSocket 服务 (带自愈守护循环，绝不因单一错误崩溃退出)
	srv, err := server.NewServer(cfg)
	if err != nil {
		log.Fatalf("初始化 HTTP 服务失败: %v", err)
	}

	go func() {
		for {
			if err := srv.Start(); err != nil {
				log.Printf("[TimAgent 守护] HTTP 服务连接断开: %v，1秒后自动恢复监听...", err)
				time.Sleep(1 * time.Second)
			} else {
				return
			}
		}
	}()

	// 启动内部端口探活看门狗，确保外部端口和本地面板始终可以访问
	watchdog := daemon.StartInternalWatchdog(
		func() (int, int) {
			return srv.PublicPort(), srv.AdminPort()
		},
		func() {
			log.Printf("[TimAgent 守护] 看门狗探测到端口异常，执行自愈重启...")
			_ = srv.Close()
			go func() {
				_ = srv.Start()
			}()
		},
	)
	defer watchdog.Stop()

	// 等待本地控制面板端口就绪
	for i := 0; i < 50; i++ {
		if srv.AdminPort() > 0 {
			break
		}
		time.Sleep(20 * time.Millisecond)
	}

	adminPort := srv.AdminPort()
	if adminPort == 0 {
		adminPort = currentCfg.Port
	}

	// 5. 自动弹出浏览器访问本机控制面板 (无需密码直接进入)
	if currentCfg.AutoOpen && !*flagNoOpen {
		go func() {
			time.Sleep(200 * time.Millisecond)
			dashURL := fmt.Sprintf("http://127.0.0.1:%d/", adminPort)
			_ = tray.OpenURL(dashURL)
		}()
	}

	// 6. 运行常驻系统托盘
	// 注册配置变更监听器：当在控制面板切换语言时，自动同步托盘菜单文案
	cfg.AddListener(func(d config.ConfigData) {
		tray.UpdateTrayLanguage(d.Language)
	})

	trayCallbacks := tray.TrayCallbacks{
		Port:     currentCfg.Port,
		Language: currentCfg.Language,
		OnLanguageChange: func(newLang string) {
			c := cfg.Get()
			if c.Language != newLang {
				_ = cfg.Update(c.Port, c.Password, c.Shell, c.WorkDir, c.SelectedAgent, c.Theme, newLang, c.AutoApprove, c.KeepAlive, nil)
				log.Printf("[TimAgent] 语言已通过托盘切换为: %s\n", newLang)
			}
		},
		OnOpenGUI: func() {
			_ = tray.OpenURL(fmt.Sprintf("http://127.0.0.1:%d/", srv.AdminPort()))
		},
		OnOpenTerm: func() {
			termURL := fmt.Sprintf("http://127.0.0.1:%d/terminal?token=%s", currentCfg.Port, currentCfg.AuthToken)
			_ = tray.OpenURL(termURL)
		},
		OnRestartPty: func() {
			c := cfg.Get()
			agentCmd := cfg.GetAgentCommand(c.SelectedAgent, c.AutoApprove)
			_, _ = mgr.Restart("default", "主会话", c.Shell, c.WorkDir, c.SelectedAgent, agentCmd, c.Language)
			log.Println("[TimAgent] 终端主会话已重启")
		},
		OnExit: func() {
			log.Println("[TimAgent] 正在退出服务...")
			watchdog.Stop()
			_ = srv.Close()
			mgr.CloseAll()
			os.Exit(0)
		},
	}

	tray.RunTray(trayCallbacks)
}
