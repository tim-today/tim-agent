package daemon

import (
	"encoding/json"
	"fmt"
	"log"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"time"
)

const (
	EnvWorkerMarker  = "TIM_AGENT_WORKER_PROC"
	EnvRespawnMarker = "TIM_AGENT_RESPAWN"
)

// InstanceInfo 记录当前正在运行的单实例进程与端口信息
type InstanceInfo struct {
	PID        int    `json:"pid"`
	AdminPort  int    `json:"admin_port"`
	PublicPort int    `json:"public_port"`
	StartedAt  string `json:"started_at"`
}

func stateDir() string {
	home, err := os.UserHomeDir()
	if err != nil || home == "" {
		return os.TempDir()
	}
	dir := filepath.Join(home, ".tim-agent")
	_ = os.MkdirAll(dir, 0755)
	return dir
}

func instanceFilePath() string {
	return filepath.Join(stateDir(), "instance.json")
}

func intentionalExitFilePath() string {
	return filepath.Join(stateDir(), "intentional_exit")
}

func workerLockFilePath() string {
	return filepath.Join(stateDir(), "worker.lock")
}

func supervisorLockFilePath() string {
	return filepath.Join(stateDir(), "supervisor.lock")
}

// IsWorkerProcess 检查当前进程是否是由守护 Supervisor 启动的子进程
func IsWorkerProcess() bool {
	return os.Getenv(EnvWorkerMarker) == "1"
}

// IsRespawnedProcess 检查当前进程是否是异常退出后被守护进程自动拉起恢复的实例
func IsRespawnedProcess() bool {
	return os.Getenv(EnvRespawnMarker) == "1"
}

// AcquireSingleInstanceLock 获取工作进程全局单实例文件锁，确保全系统绝不多开或覆盖已有进程
func AcquireSingleInstanceLock() (*os.File, bool) {
	f, err := os.OpenFile(workerLockFilePath(), os.O_CREATE|os.O_RDWR, 0600)
	if err != nil {
		return nil, true
	}
	if !TryLockFile(f) {
		_ = f.Close()
		return nil, false
	}
	return f, true
}

// SaveInstanceInfo 写入当前运行实例的 PID 与固定端口信息
func SaveInstanceInfo(adminPort, publicPort int) {
	info := InstanceInfo{
		PID:        os.Getpid(),
		AdminPort:  adminPort,
		PublicPort: publicPort,
		StartedAt:  time.Now().Format(time.RFC3339),
	}
	if data, err := json.MarshalIndent(info, "", "  "); err == nil {
		_ = os.WriteFile(instanceFilePath(), data, 0600)
	}
}

// LoadInstanceInfo 读取已记录的运行实例信息
func LoadInstanceInfo() (*InstanceInfo, error) {
	data, err := os.ReadFile(instanceFilePath())
	if err != nil {
		return nil, err
	}
	var info InstanceInfo
	if err := json.Unmarshal(data, &info); err != nil {
		return nil, err
	}
	return &info, nil
}

// ClearInstanceInfo 清理运行实例记录
func ClearInstanceInfo() {
	_ = os.Remove(instanceFilePath())
}

// MarkIntentionalExit 标记用户通过托盘菜单主动手动退出（仅手动退出时不再自动拉起）
func MarkIntentionalExit() {
	_ = os.WriteFile(intentionalExitFilePath(), []byte(fmt.Sprintf("%d", os.Getpid())), 0600)
	ClearInstanceInfo()
}

// ClearIntentionalExit 清除手动退出标记
func ClearIntentionalExit() {
	_ = os.Remove(intentionalExitFilePath())
}

// ConsumeIntentionalExit 检查并消费手动退出标记
func ConsumeIntentionalExit() bool {
	path := intentionalExitFilePath()
	if _, err := os.Stat(path); err == nil {
		_ = os.Remove(path)
		return true
	}
	return false
}

// RunSupervisor 启动外部进程守护监控器
// 默认常驻保护主工作进程：非用户手动退出（如异常闪退、崩溃、意外中止）时自动在 1 秒内恢复重启，保持原端口与服务不中断
func RunSupervisor() {
	IgnoreHangupSignals()

	supLock, err := os.OpenFile(supervisorLockFilePath(), os.O_CREATE|os.O_RDWR, 0600)
	if err == nil {
		if !TryLockFile(supLock) {
			_ = supLock.Close()
			log.Println("[Supervisor 守护进程] 检测到已有守护进程在运行，当前重复守护进程直接退出。")
			return
		}
		defer func() {
			UnlockFile(supLock)
			_ = supLock.Close()
		}()
	}

	ClearIntentionalExit()

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

	log.Println("[Supervisor 守护进程] 已就绪，正在守护 Tim-Agent 主服务 (非手动退出将自动重启恢复)...")

	isRespawn := false
	for {
		cmd := exec.Command(executable, childArgs...)
		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		env := append(os.Environ(), EnvWorkerMarker+"=1")
		if isRespawn {
			env = append(env, EnvRespawnMarker+"=1")
		}
		cmd.Env = env

		startTime := time.Now()
		if err := cmd.Start(); err != nil {
			log.Printf("[Supervisor 守护进程] 启动工作子进程失败: %v，将在 1 秒后重试...", err)
			time.Sleep(1 * time.Second)
			isRespawn = true
			continue
		}

		childPid := cmd.Process.Pid
		log.Printf("[Supervisor 守护进程] 工作进程已拉起 (PID: %d)", childPid)

		waitErr := cmd.Wait()
		duration := time.Since(startTime)

		// 若用户在托盘点击了「退出服务」，则正常终止守护进程
		if ConsumeIntentionalExit() {
			log.Printf("[Supervisor 守护进程] 收到用户手动退出指令，服务已安全停止 (PID: %d)。", childPid)
			return
		}

		log.Printf("[Supervisor 守护进程] 警告: 工作进程 (PID: %d) 非手动异常退出 (运行耗时: %v, 原因: %v)", childPid, duration, waitErr)
		isRespawn = true

		if duration < 2*time.Second {
			log.Println("[Supervisor 守护进程] 进程退出过快，等待 1.5 秒后自动恢复重启...")
			time.Sleep(1500 * time.Millisecond)
		} else {
			log.Println("[Supervisor 守护进程] 正在立即重新拉起服务，确保原端口与会话持续可用...")
			time.Sleep(600 * time.Millisecond)
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
