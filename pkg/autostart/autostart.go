package autostart

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

const (
	macLaunchAgentLabel = "com.tim-agent.launcher"
	macPlistTemplate    = `<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
    <key>Label</key>
    <string>%s</string>
    <key>ProgramArguments</key>
    <array>
        <string>%s</string>
    </array>
    <key>RunAtLoad</key>
    <true/>
    <key>KeepAlive</key>
    <false/>
    <key>ProcessType</key>
    <string>Interactive</string>
</dict>
</plist>
`
	linuxDesktopTemplate = `[Desktop Entry]
Type=Application
Version=1.0
Name=TimAgent
Comment=Tim-Agent Terminal & Mobile Vibe Coding Hub
Exec="%s"
Terminal=false
StartupNotify=false
X-GNOME-Autostart-enabled=true
`
)

// getExecutablePath 获取最佳自启动目标可执行文件路径
func getExecutablePath() string {
	homeDir, _ := os.UserHomeDir()

	switch runtime.GOOS {
	case "darwin":
		candidates := []string{
			"/Applications/TimAgent.app/Contents/MacOS/tim-agent",
			filepath.Join(homeDir, "Applications", "TimAgent.app", "Contents", "MacOS", "tim-agent"),
		}
		for _, cand := range candidates {
			if info, err := os.Stat(cand); err == nil && !info.IsDir() {
				return cand
			}
		}
	case "windows":
		if homeDir != "" {
			cand := filepath.Join(homeDir, ".tim-agent", "TimAgent.exe")
			if info, err := os.Stat(cand); err == nil && !info.IsDir() {
				return cand
			}
		}
	default: // linux
		if homeDir != "" {
			cand := filepath.Join(homeDir, ".local", "bin", "tim-agent")
			if info, err := os.Stat(cand); err == nil && !info.IsDir() {
				return cand
			}
		}
	}

	exe, err := os.Executable()
	if err == nil {
		if resolved, err := filepath.EvalSymlinks(exe); err == nil {
			return resolved
		}
		return exe
	}
	return "tim-agent"
}

// getMacPlistPath 获取 macOS LaunchAgents plist 文件路径
func getMacPlistPath() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(homeDir, "Library", "LaunchAgents", fmt.Sprintf("%s.plist", macLaunchAgentLabel)), nil
}

// getLinuxAutostartPath 获取 Linux autostart desktop 文件路径
func getLinuxAutostartPath() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(homeDir, ".config", "autostart", "tim-agent.desktop"), nil
}

// IsEnabled 检查系统当前是否已开启开机自动启动
func IsEnabled() bool {
	switch runtime.GOOS {
	case "darwin":
		plistPath, err := getMacPlistPath()
		if err != nil {
			return false
		}
		_, err = os.Stat(plistPath)
		return err == nil

	case "windows":
		cmd := exec.Command("reg", "query", `HKCU\Software\Microsoft\Windows\CurrentVersion\Run`, "/v", "TimAgent")
		return cmd.Run() == nil

	default: // linux
		desktopPath, err := getLinuxAutostartPath()
		if err != nil {
			return false
		}
		_, err = os.Stat(desktopPath)
		return err == nil
	}
}

// Enable 开启开机自动启动服务
func Enable() error {
	exePath := getExecutablePath()

	switch runtime.GOOS {
	case "darwin":
		plistPath, err := getMacPlistPath()
		if err != nil {
			return err
		}
		agentDir := filepath.Dir(plistPath)
		if err := os.MkdirAll(agentDir, 0755); err != nil {
			return fmt.Errorf("failed to create LaunchAgents directory: %w", err)
		}
		content := fmt.Sprintf(macPlistTemplate, macLaunchAgentLabel, exePath)
		if err := os.WriteFile(plistPath, []byte(content), 0644); err != nil {
			return fmt.Errorf("failed to write LaunchAgent plist: %w", err)
		}
		// 尝试使用 launchctl load 注册
		_ = exec.Command("launchctl", "unload", "-w", plistPath).Run()
		_ = exec.Command("launchctl", "load", "-w", plistPath).Run()
		return nil

	case "windows":
		val := fmt.Sprintf(`"%s"`, exePath)
		cmd := exec.Command("reg", "add", `HKCU\Software\Microsoft\Windows\CurrentVersion\Run`, "/v", "TimAgent", "/t", "REG_SZ", "/d", val, "/f")
		if out, err := cmd.CombinedOutput(); err != nil {
			return fmt.Errorf("failed to add registry run key: %v (%s)", err, strings.TrimSpace(string(out)))
		}
		return nil

	default: // linux
		desktopPath, err := getLinuxAutostartPath()
		if err != nil {
			return err
		}
		autostartDir := filepath.Dir(desktopPath)
		if err := os.MkdirAll(autostartDir, 0755); err != nil {
			return fmt.Errorf("failed to create autostart directory: %w", err)
		}
		content := fmt.Sprintf(linuxDesktopTemplate, exePath)
		if err := os.WriteFile(desktopPath, []byte(content), 0755); err != nil {
			return fmt.Errorf("failed to write autostart desktop file: %w", err)
		}
		return nil
	}
}

// Disable 禁用并移除开机自动启动服务
func Disable() error {
	switch runtime.GOOS {
	case "darwin":
		plistPath, err := getMacPlistPath()
		if err != nil {
			return err
		}
		_ = exec.Command("launchctl", "unload", "-w", plistPath).Run()
		if err := os.Remove(plistPath); err != nil && !os.IsNotExist(err) {
			return fmt.Errorf("failed to remove LaunchAgent plist: %w", err)
		}
		return nil

	case "windows":
		cmd := exec.Command("reg", "delete", `HKCU\Software\Microsoft\Windows\CurrentVersion\Run`, "/v", "TimAgent", "/f")
		_ = cmd.Run()
		return nil

	default: // linux
		desktopPath, err := getLinuxAutostartPath()
		if err != nil {
			return err
		}
		if err := os.Remove(desktopPath); err != nil && !os.IsNotExist(err) {
			return fmt.Errorf("failed to remove autostart desktop file: %w", err)
		}
		return nil
	}
}

// Set 根据布尔值设置开机自启开关状态
func Set(enable bool) error {
	if enable {
		return Enable()
	}
	return Disable()
}
