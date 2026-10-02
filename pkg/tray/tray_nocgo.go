//go:build nocgo || (!darwin && !windows && !linux)

package tray

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
)

type TrayCallbacks struct {
	Port             int
	Language         string
	OnLanguageChange func(string)
	OnOpenGUI        func()
	OnOpenTerm       func()
	OnRestartPty     func()
	OnExit           func()
}

// UpdateTrayLanguage nocgo 桩函数
func UpdateTrayLanguage(lang string) {}

func RunTray(callbacks TrayCallbacks) {
	fmt.Printf("[TimAgent] 服务已在端口 :%d 启动，正在后台监听...\n", callbacks.Port)
	fmt.Println("按 Ctrl+C 退出程序")

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)

	<-sigCh
	if callbacks.OnExit != nil {
		callbacks.OnExit()
	}
}
