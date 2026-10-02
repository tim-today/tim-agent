//go:build windows

package daemon

// IgnoreHangupSignals Windows 平台桩函数
func IgnoreHangupSignals() {}
