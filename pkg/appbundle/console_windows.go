//go:build windows

package appbundle

import (
	"os"
	"syscall"
	"unsafe"
)

const (
	attachParentProcess = ^uint32(0) // (DWORD)-1
	swHide              = 0
)

// AdaptWindowsConsole 智能适配 Windows 鼠标双击启动与命令行启动：
// 1. 如果是 -H=windowsgui 编译的 GUI 程序在命令行中带参数运行，自动挂载父终端输出；
// 2. 如果是普通编译的 .exe 被用户用鼠标在资源管理器直接双击打开（独立控制台进程数==1且无命令行参数），
//    自动隐藏黑框控制台窗口，转为纯净后台托盘 GUI 模式运行。
func AdaptWindowsConsole() {
	kernel32 := syscall.NewLazyDLL("kernel32.dll")
	procGetConsoleWindow := kernel32.NewProc("GetConsoleWindow")
	procGetConsoleProcessList := kernel32.NewProc("GetConsoleProcessList")
	procAttachConsole := kernel32.NewProc("AttachConsole")
	procFreeConsole := kernel32.NewProc("FreeConsole")

	hwnd, _, _ := procGetConsoleWindow.Call()
	if hwnd == 0 {
		// 当前为 -H=windowsgui 无控制台模式：若带有命令行参数，尝试附加到父级命令行窗口
		if len(os.Args) > 1 {
			_, _, _ = procAttachConsole.Call(uintptr(attachParentProcess))
		}
		return
	}

	// 当前有控制台窗口：检查是否由资源管理器鼠标双击直接启动（仅当前1个进程附着且无额外命令行参数）
	if len(os.Args) == 1 {
		var pids [4]uint32
		count, _, _ := procGetConsoleProcessList.Call(uintptr(unsafe.Pointer(&pids[0])), 4)
		if count <= 1 {
			user32 := syscall.NewLazyDLL("user32.dll")
			procShowWindow := user32.NewProc("ShowWindow")
			_, _, _ = procShowWindow.Call(hwnd, uintptr(swHide))
			_, _, _ = procFreeConsole.Call()
		}
	}
}
