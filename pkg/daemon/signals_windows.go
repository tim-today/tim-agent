//go:build windows

package daemon

import "os"

// IgnoreHangupSignals Windows 平台桩函数
func IgnoreHangupSignals() {}

// TryLockFile Windows 平台通过单实例文件与端口双重校验
func TryLockFile(f *os.File) bool {
	return f != nil
}

// UnlockFile 释放文件排他锁
func UnlockFile(f *os.File) {}

// IsProcessAlive 检测指定 PID 的进程是否存在
func IsProcessAlive(pid int) bool {
	if pid <= 0 {
		return false
	}
	p, err := os.FindProcess(pid)
	return err == nil && p != nil
}
