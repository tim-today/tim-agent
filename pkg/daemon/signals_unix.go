//go:build !windows

package daemon

import (
	"log"
	"os"
	"os/signal"
	"syscall"
)

// IgnoreHangupSignals 免疫挂断信号和管道中断信号
// 在 macOS / Linux 下，当用户关闭启动该程序的 Terminal 窗口时，系统会发送 SIGHUP。
// 忽略此信号可确保关闭终端窗口或终端会话结束时，服务依然在后台稳固运行，绝不闪退。
func IgnoreHangupSignals() {
	sigCh := make(chan os.Signal, 10)
	signal.Notify(sigCh, syscall.SIGHUP, syscall.SIGPIPE, syscall.SIGUSR1, syscall.SIGUSR2)
	go func() {
		for sig := range sigCh {
			log.Printf("[Daemon] 收到系统信号: %v (已自动拦截并免疫，服务继续保持常驻运行)", sig)
		}
	}()
}

// TryLockFile 尝试对文件施加非阻塞排他锁，防止多实例并发启动
func TryLockFile(f *os.File) bool {
	if f == nil {
		return false
	}
	err := syscall.Flock(int(f.Fd()), syscall.LOCK_EX|syscall.LOCK_NB)
	return err == nil
}

// UnlockFile 释放文件排他锁
func UnlockFile(f *os.File) {
	if f == nil {
		return
	}
	_ = syscall.Flock(int(f.Fd()), syscall.LOCK_UN)
}

// IsProcessAlive 检测指定 PID 的进程是否仍在运行
func IsProcessAlive(pid int) bool {
	if pid <= 0 {
		return false
	}
	err := syscall.Kill(pid, 0)
	return err == nil
}
