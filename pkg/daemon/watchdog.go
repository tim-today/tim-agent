package daemon

import (
	"fmt"
	"log"
	"net"
	"os"
	"os/exec"
	"os/signal"
	"syscall"
	"time"
)

const EnvWorkerMarker = "TIM_AGENT_WORKER_PROC"

// IsWorkerProcess 检查当前进程是否是由守护 Supervisor 启动的子进程
func IsWorkerProcess() bool {
	return os.Getenv(EnvWorkerMarker) == "1"
}

// RunSupervisor 启动外部进程守护监控器
// 监控工作子进程，若子进程异常闪退或被杀死，Supervisor 在 1 秒内自动重启它，保证端口持续常驻可用
func RunSupervisor() {
	IgnoreHangupSignals()

	// 捕获系统退出信号（Ctrl+C 或 kill），向子进程转发退出
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)

	executable, err := os.Executable()
	if err != nil {
		executable = os.Args[0]
	}

	args := os.Args[1:]
	// 去除 --daemon 或 -daemon 参数，避免无限递归
	var childArgs []string
	for _, arg := range args {
		if arg != "-daemon" && arg != "--daemon" {
			childArgs = append(childArgs, arg)
		}
	}

	log.Println("[Supervisor 守护进程] 已就绪，开始守护 Tim-Agent 主服务...")

	for {
		cmd := exec.Command(executable, childArgs...)
		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		cmd.Env = append(os.Environ(), EnvWorkerMarker+"=1")

		startTime := time.Now()
		if err := cmd.Start(); err != nil {
			log.Printf("[Supervisor 守护进程] 启动工作子进程失败: %v，将在 2 秒后重试...", err)
			time.Sleep(2 * time.Second)
			continue
		}

		childPid := cmd.Process.Pid
		log.Printf("[Supervisor 守护进程] 工作进程已拉起 (PID: %d)", childPid)

		// 监听父进程退出信号并妥善处理
		exitChan := make(chan error, 1)
		go func() {
			exitChan <- cmd.Wait()
		}()

		select {
		case sig := <-sigCh:
			log.Printf("[Supervisor 守护进程] 收到退出信号 (%v)，正在停止工作进程...", sig)
			if cmd.Process != nil {
				_ = cmd.Process.Signal(sig)
				time.Sleep(500 * time.Millisecond)
				_ = cmd.Process.Kill()
			}
			os.Exit(0)
		case waitErr := <-exitChan:
			duration := time.Since(startTime)
			log.Printf("[Supervisor 守护进程] 警告: 工作进程 (PID: %d) 意外退出 (运行耗时: %v, 原因: %v)", childPid, duration, waitErr)
			
			// 如果运行时间太短（不足 1 秒就崩溃），避免剧烈死循环，稍作休眠
			if duration < 2*time.Second {
				log.Println("[Supervisor 守护进程] 进程退出过快，等待 2 秒后重启...")
				time.Sleep(2 * time.Second)
			} else {
				log.Println("[Supervisor 守护进程] 正在立即重新拉起服务，确保端口继续可用...")
				time.Sleep(500 * time.Millisecond)
			}
		}
	}
}

// WatchdogPortChecker 内部看门狗：周期性检测端口可用性，异常时自愈重试
type WatchdogPortChecker struct {
	targetPort int
	stopCh     chan struct{}
}

// StartInternalWatchdog 启动内部端口健康看门狗
func StartInternalWatchdog(portGetter func() (int, int), onPortDown func()) *WatchdogPortChecker {
	w := &WatchdogPortChecker{
		stopCh: make(chan struct{}),
	}

	go func() {
		defer func() {
			if r := recover(); r != nil {
				log.Printf("[Watchdog 看门狗] 捕获异常: %v", r)
			}
		}()

		ticker := time.NewTicker(4 * time.Second)
		defer ticker.Stop()

		failCount := 0

		for {
			select {
			case <-w.stopCh:
				return
			case <-ticker.C:
				pubPort, adminPort := portGetter()
				if pubPort <= 0 {
					continue
				}

				// 检测公网共享端口是否正常监听
				pubOK := checkTCPPort(pubPort)
				adminOK := true
				if adminPort > 0 {
					adminOK = checkTCPPort(adminPort)
				}

				if !pubOK || !adminOK {
					failCount++
					log.Printf("[Watchdog 看门狗] 端口探活异常 (Public:%d -> %v, Admin:%d -> %v, 连续失败: %d/2)",
						pubPort, pubOK, adminPort, adminOK, failCount)

					if failCount >= 2 {
						log.Printf("[Watchdog 看门狗] 端口连续两次探活失败，触发服务自愈重连...")
						failCount = 0
						if onPortDown != nil {
							onPortDown()
						}
					}
				} else {
					failCount = 0
				}
			}
		}
	}()

	return w
}

func (w *WatchdogPortChecker) Stop() {
	close(w.stopCh)
}

func checkTCPPort(port int) bool {
	conn, err := net.DialTimeout("tcp", fmt.Sprintf("127.0.0.1:%d", port), 800*time.Millisecond)
	if err != nil {
		return false
	}
	_ = conn.Close()
	return true
}
