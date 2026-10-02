package tray

import (
	"os/exec"
	"runtime"
)

// OpenURL 跨平台自动打开浏览器访问指定网址
func OpenURL(url string) error {
	var cmd *exec.Cmd

	switch runtime.GOOS {
	case "darwin":
		cmd = exec.Command("open", url)
	case "windows":
		cmd = exec.Command("cmd", "/c", "start", url)
	default: // linux / bsd
		cmd = exec.Command("xdg-open", url)
	}

	return cmd.Start()
}
