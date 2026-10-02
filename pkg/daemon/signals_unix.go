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
	signal.Notify(sigCh, syscall.SIGHUP, syscall.SIGPIPE)
	go func() {
		for sig := range sigCh {
			log.Printf("[Daemon] 收到系统信号: %v (已自动拦截并免疫，服务继续保持常驻运行)", sig)
		}
	}()
}
