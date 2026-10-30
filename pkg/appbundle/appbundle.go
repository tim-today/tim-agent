package appbundle

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/tim-today/tim-agent/pkg/config"
	"github.com/tim-today/tim-agent/pkg/tray"
)

// InitGUIEnvironment 初始化鼠标双击 GUI 启动时的运行环境：
// 1. Windows 自动适配无黑框 GUI 控制台模式；
// 2. macOS / Linux 双击 .app 或 .desktop 启动时补全完整 PATH 环境变量（Homebrew / npm / nvm / local bin 等），确保能找到 claude/codex 等 CLI；
// 3. 当由 Finder / 桌面启动器在根目录 "/" 或系统目录启动时，自动将默认工作目录切回用户主目录；
// 4. 挂载 ~/.tim-agent/tim-agent.log 日志输出，便于无终端窗口时记录运行状态。
func InitGUIEnvironment() {
	AdaptWindowsConsole()

	homeDir, err := os.UserHomeDir()
	if err != nil || homeDir == "" {
		return
	}

	// 1. 若当前工作目录是 "/" (macOS Finder 双击 .app 默认目录) 或 Windows System32，自动切换到用户主目录
	if cwd, err := os.Getwd(); err == nil {
		if cwd == "/" || strings.EqualFold(cwd, `C:\Windows\System32`) {
			_ = os.Chdir(homeDir)
		}
	}

	// 2. 补全 GUI 模式下的 PATH 环境变量
	if runtime.GOOS != "windows" {
		extraPaths := []string{
			"/opt/homebrew/bin",
			"/opt/homebrew/sbin",
			"/usr/local/bin",
			"/usr/local/sbin",
			filepath.Join(homeDir, ".local", "bin"),
			filepath.Join(homeDir, ".bun", "bin"),
			filepath.Join(homeDir, ".cargo", "bin"),
			filepath.Join(homeDir, "go", "bin"),
			filepath.Join(homeDir, ".npm-global", "bin"),
		}

		// 自动探测 ~/.nvm/versions/node/*/bin 与 ~/.fnm
		if nvmGlob, err := filepath.Glob(filepath.Join(homeDir, ".nvm", "versions", "node", "*", "bin")); err == nil {
			for i := len(nvmGlob) - 1; i >= 0; i-- {
				extraPaths = append(extraPaths, nvmGlob[i])
			}
		}

		currentPath := os.Getenv("PATH")
		pathSet := make(map[string]bool)
		var merged []string
		for _, p := range strings.Split(currentPath, string(os.PathListSeparator)) {
			if p != "" && !pathSet[p] {
				pathSet[p] = true
				merged = append(merged, p)
			}
		}
		for _, p := range extraPaths {
			if !pathSet[p] {
				if info, err := os.Stat(p); err == nil && info.IsDir() {
					pathSet[p] = true
					merged = append(merged, p)
				}
			}
		}
		_ = os.Setenv("PATH", strings.Join(merged, string(os.PathListSeparator)))
	}

	// 3. 挂载持久化日志文件 ~/.tim-agent/tim-agent.log
	logDir := filepath.Join(homeDir, ".tim-agent")
	if err := os.MkdirAll(logDir, 0755); err == nil {
		logPath := filepath.Join(logDir, "tim-agent.log")
		if logFile, err := os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644); err == nil {
			log.SetOutput(io.MultiWriter(os.Stderr, logFile))
		}
	}
}

// WakeExistingInstance 检测本机是否已有正在运行的 Tim-Agent 实例。
// 若已有实例运行，则直接自动弹出其控制面板网页并返回 true（防止用户多次双击桌面图标拉起重复托盘）。
func WakeExistingInstance(publicPort int) bool {
	client := &http.Client{Timeout: 350 * time.Millisecond}

	// 优先探测本地免密控制面板端口 (20997 ~ 21002)
	for port := 20997; port <= 21002; port++ {
		statusURL := fmt.Sprintf("http://127.0.0.1:%d/api/status", port)
		resp, err := client.Get(statusURL)
		if err != nil {
			continue
		}
		var data map[string]any
		decodeErr := json.NewDecoder(resp.Body).Decode(&data)
		_ = resp.Body.Close()
		if decodeErr == nil && data["version"] != nil && data["config"] != nil {
			dashURL := fmt.Sprintf("http://127.0.0.1:%d/", port)
			log.Printf("[TimAgent] 检测到本机实例已在运行，直接唤醒控制面板: %s", dashURL)
			_ = tray.OpenURL(dashURL)
			return true
		}
	}

	// 备用探测配置的共享端口
	if publicPort > 0 {
		statusURL := fmt.Sprintf("http://127.0.0.1:%d/api/status", publicPort)
		if resp, err := client.Get(statusURL); err == nil {
			var data map[string]any
			decodeErr := json.NewDecoder(resp.Body).Decode(&data)
			_ = resp.Body.Close()
			if decodeErr == nil && data["version"] != nil {
				dashURL := fmt.Sprintf("http://127.0.0.1:%d/", publicPort)
				log.Printf("[TimAgent] 检测到本机实例已在端口 %d 运行，唤醒控制面板: %s", publicPort, dashURL)
				_ = tray.OpenURL(dashURL)
				return true
			}
		}
	}

	return false
}

// PackageMacApp 将指定二进制打包为标准 macOS GUI 应用程序包 (TimAgent.app)
func PackageMacApp(binaryPath, outAppDir string) error {
	if outAppDir == "" {
		outAppDir = "TimAgent.app"
	}
	contentsDir := filepath.Join(outAppDir, "Contents")
	macOSDir := filepath.Join(contentsDir, "MacOS")
	resourcesDir := filepath.Join(contentsDir, "Resources")

	if err := os.MkdirAll(macOSDir, 0755); err != nil {
		return fmt.Errorf("创建 MacOS 目录失败: %w", err)
	}
	if err := os.MkdirAll(resourcesDir, 0755); err != nil {
		return fmt.Errorf("创建 Resources 目录失败: %w", err)
	}

	// 1. 复制可执行程序至 Contents/MacOS/tim-agent
	targetBin := filepath.Join(macOSDir, "tim-agent")
	if err := copyFile(binaryPath, targetBin, 0755); err != nil {
		return fmt.Errorf("复制可执行文件失败: %w", err)
	}

	// 2. 生成高清图标 Contents/Resources/AppIcon.icns
	icnsPath := filepath.Join(resourcesDir, "AppIcon.icns")
	if err := os.WriteFile(icnsPath, tray.GenerateICNS(), 0644); err != nil {
		return fmt.Errorf("写入 AppIcon.icns 失败: %w", err)
	}

	// 3. 写入 Contents/Info.plist
	infoPlist := fmt.Sprintf(`<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
	<key>CFBundleDevelopmentRegion</key>
	<string>zh_CN</string>
	<key>CFBundleDisplayName</key>
	<string>TimAgent</string>
	<key>CFBundleExecutable</key>
	<string>tim-agent</string>
	<key>CFBundleIconFile</key>
	<string>AppIcon</string>
	<key>CFBundleIdentifier</key>
	<string>com.timtoday.tim-agent</string>
	<key>CFBundleInfoDictionaryVersion</key>
	<string>6.0</string>
	<key>CFBundleName</key>
	<string>TimAgent</string>
	<key>CFBundlePackageType</key>
	<string>APPL</string>
	<key>CFBundleShortVersionString</key>
	<string>%s</string>
	<key>CFBundleVersion</key>
	<string>%s</string>
	<key>LSMinimumSystemVersion</key>
	<string>11.0</string>
	<key>LSArchitecturePriority</key>
	<array>
		<string>arm64</string>
		<string>x86_64</string>
	</array>
	<key>NSHighResolutionCapable</key>
	<true/>
	<key>LSUIElement</key>
	<false/>
</dict>
</plist>
`, strings.TrimPrefix(config.AppVersion, "v"), strings.TrimPrefix(config.AppVersion, "v"))

	plistPath := filepath.Join(contentsDir, "Info.plist")
	if err := os.WriteFile(plistPath, []byte(infoPlist), 0644); err != nil {
		return fmt.Errorf("写入 Info.plist 失败: %w", err)
	}

	// 如果当前在 macOS 上，移除 quarantine 属性、进行 ad-hoc 签名并刷新 LaunchServices 缓存以支持直接双击打开
	if runtime.GOOS == "darwin" {
		_ = exec.Command("xattr", "-cr", outAppDir).Run()
		_ = exec.Command("codesign", "--force", "--deep", "--sign", "-", outAppDir).Run()
		lsreg := "/System/Library/Frameworks/CoreServices.framework/Frameworks/LaunchServices.framework/Support/lsregister"
		if _, err := os.Stat(lsreg); err == nil {
			_ = exec.Command(lsreg, "-f", outAppDir).Run()
		}
	}

	return nil
}

// PackageLinuxBundle 将 Linux 二进制打包为包含 .desktop 桌面图标启动器与一键安装脚本的 GUI 应用目录
func PackageLinuxBundle(binaryPath, outDir string) error {
	if err := os.MkdirAll(outDir, 0755); err != nil {
		return err
	}

	targetBin := filepath.Join(outDir, "tim-agent")
	if err := copyFile(binaryPath, targetBin, 0755); err != nil {
		return err
	}

	iconPath := filepath.Join(outDir, "tim-agent.png")
	if err := os.WriteFile(iconPath, tray.GenerateTAIcon(256), 0644); err != nil {
		return err
	}

	// 支持双击直接运行的极简 AppRun / 启动器脚本
	launcherScript := `#!/usr/bin/env bash
set -e
APP_DIR="$(cd "$(dirname "$0")" && pwd)"
chmod +x "$APP_DIR/tim-agent"

# 自动注册桌面与系统菜单应用图标
mkdir -p "$HOME/.local/share/applications" "$HOME/.local/share/icons"
cp -f "$APP_DIR/tim-agent.png" "$HOME/.local/share/icons/tim-agent.png"

cat > "$HOME/.local/share/applications/tim-agent.desktop" <<EOF
[Desktop Entry]
Version=1.0
Type=Application
Name=TimAgent
Comment=Tim-Agent Terminal & Mobile Vibe Coding Hub
Exec="$APP_DIR/tim-agent"
Icon=$HOME/.local/share/icons/tim-agent.png
Terminal=false
Categories=Development;Utility;TerminalEmulator;
StartupNotify=false
EOF
chmod +x "$HOME/.local/share/applications/tim-agent.desktop"

if [ -d "$HOME/Desktop" ]; then
  cp -f "$HOME/.local/share/applications/tim-agent.desktop" "$HOME/Desktop/TimAgent.desktop"
  chmod +x "$HOME/Desktop/TimAgent.desktop"
  command -v gio >/dev/null 2>&1 && gio set "$HOME/Desktop/TimAgent.desktop" metadata::trusted true 2>/dev/null || true
fi

nohup "$APP_DIR/tim-agent" >/dev/null 2>&1 &
echo "TimAgent GUI Application started and desktop icon installed!"
`
	if err := os.WriteFile(filepath.Join(outDir, "Start-TimAgent-GUI.sh"), []byte(launcherScript), 0755); err != nil {
		return err
	}

	desktopTemplate := `[Desktop Entry]
Version=1.0
Type=Application
Name=TimAgent
Comment=Tim-Agent Terminal & Mobile Vibe Coding Hub
Exec= bash -c '"$(dirname "$1")/tim-agent"' dummy %k
Icon=utilities-terminal
Terminal=false
Categories=Development;Utility;
`
	if err := os.WriteFile(filepath.Join(outDir, "TimAgent.desktop"), []byte(desktopTemplate), 0755); err != nil {
		return err
	}

	return nil
}

// InstallDesktopApp 将当前正在运行的程序一键生成并安装为本机操作系统的桌面 GUI 应用程序（支持鼠标直接点开）
func InstallDesktopApp() (string, error) {
	exePath, err := os.Executable()
	if err != nil {
		return "", fmt.Errorf("获取当前程序路径失败: %w", err)
	}
	if resolved, err := filepath.EvalSymlinks(exePath); err == nil {
		exePath = resolved
	}

	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("获取用户主目录失败: %w", err)
	}

	switch runtime.GOOS {
	case "darwin":
		// 优先尝试安装到 /Applications/TimAgent.app，若无权限则安装到 ~/Applications/TimAgent.app，同时在桌面生成快捷应用
		sysAppPath := "/Applications/TimAgent.app"
		if err := PackageMacApp(exePath, sysAppPath); err != nil {
			userAppDir := filepath.Join(homeDir, "Applications")
			_ = os.MkdirAll(userAppDir, 0755)
			sysAppPath = filepath.Join(userAppDir, "TimAgent.app")
			if err2 := PackageMacApp(exePath, sysAppPath); err2 != nil {
				return "", fmt.Errorf("生成 macOS 应用程序失败: %w", err2)
			}
		}
		// 同步在桌面生成一份或软链接方便直接点击
		desktopDir := filepath.Join(homeDir, "Desktop")
		if info, err := os.Stat(desktopDir); err == nil && info.IsDir() {
			desktopApp := filepath.Join(desktopDir, "TimAgent.app")
			if desktopApp != sysAppPath {
				_ = os.RemoveAll(desktopApp)
				_ = PackageMacApp(exePath, desktopApp)
			}
		}
		return sysAppPath, nil

	case "windows":
		appDataDir := filepath.Join(homeDir, ".tim-agent")
		if err := os.MkdirAll(appDataDir, 0755); err != nil {
			return "", err
		}
		// 固定拷贝一份可执行文件与 .ico 图标到 ~/.tim-agent/
		installedExe := filepath.Join(appDataDir, "TimAgent.exe")
		if !strings.EqualFold(exePath, installedExe) {
			_ = copyFile(exePath, installedExe, 0755)
		} else {
			installedExe = exePath
		}
		icoPath := filepath.Join(appDataDir, "TimAgent.ico")
		_ = os.WriteFile(icoPath, tray.GenerateICO(), 0644)

		desktopDir := filepath.Join(homeDir, "Desktop")
		desktopLnk := filepath.Join(desktopDir, "TimAgent.lnk")

		// 使用 PowerShell 创建桌面和开始菜单快捷方式 (.lnk)
		psScript := fmt.Sprintf(`
$WshShell = New-Object -ComObject WScript.Shell
$DesktopPath = [Environment]::GetFolderPath("Desktop")
$StartMenuPath = [Environment]::GetFolderPath("Programs")

foreach ($dir in @($DesktopPath, $StartMenuPath)) {
    if ($dir -and (Test-Path $dir)) {
        $lnk = $WshShell.CreateShortcut((Join-Path $dir "TimAgent.lnk"))
        $lnk.TargetPath = "%s"
        $lnk.WorkingDirectory = "%s"
        $lnk.IconLocation = "%s"
        $lnk.Description = "Tim-Agent GUI Terminal Hub"
        $lnk.Save()
    }
}
`, installedExe, homeDir, icoPath)
		cmd := exec.Command("powershell", "-NoProfile", "-NonInteractive", "-Command", psScript)
		if out, err := cmd.CombinedOutput(); err != nil {
			return "", fmt.Errorf("创建 Windows 桌面快捷方式失败: %v (%s)", err, string(out))
		}
		return desktopLnk, nil

	default: // linux
		localBinDir := filepath.Join(homeDir, ".local", "bin")
		localAppDir := filepath.Join(homeDir, ".local", "share", "applications")
		localIconDir := filepath.Join(homeDir, ".local", "share", "icons")
		_ = os.MkdirAll(localBinDir, 0755)
		_ = os.MkdirAll(localAppDir, 0755)
		_ = os.MkdirAll(localIconDir, 0755)

		installedBin := filepath.Join(localBinDir, "tim-agent")
		if exePath != installedBin {
			_ = copyFile(exePath, installedBin, 0755)
		} else {
			installedBin = exePath
		}

		iconPath := filepath.Join(localIconDir, "tim-agent.png")
		if err := os.WriteFile(iconPath, tray.GenerateTAIcon(256), 0644); err != nil {
			return "", err
		}

		desktopEntry := fmt.Sprintf(`[Desktop Entry]
Version=1.0
Type=Application
Name=TimAgent
Comment=Tim-Agent Terminal & Mobile Vibe Coding Hub
Exec="%s"
Icon=%s
Terminal=false
Categories=Development;Utility;TerminalEmulator;
StartupNotify=false
`, installedBin, iconPath)

		desktopFilePath := filepath.Join(localAppDir, "tim-agent.desktop")
		if err := os.WriteFile(desktopFilePath, []byte(desktopEntry), 0755); err != nil {
			return "", err
		}

		userDesktop := filepath.Join(homeDir, "Desktop")
		if info, err := os.Stat(userDesktop); err == nil && info.IsDir() {
			deskShortcut := filepath.Join(userDesktop, "TimAgent.desktop")
			_ = os.WriteFile(deskShortcut, []byte(desktopEntry), 0755)
			_ = exec.Command("gio", "set", deskShortcut, "metadata::trusted", "true").Run()
		}
		return desktopFilePath, nil
	}
}

func copyFile(src, dst string, perm os.FileMode) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	out, err := os.OpenFile(dst, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, perm)
	if err != nil {
		return err
	}
	defer out.Close()

	if _, err := io.Copy(out, in); err != nil {
		return err
	}
	return os.Chmod(dst, perm)
}
