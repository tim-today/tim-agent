package tray

import "fmt"

type TrayI18n struct {
	Tooltip      string
	TitleActive  string
	OpenGUI      string
	OpenGUITip   string
	OpenTerm     string
	OpenTermTip  string
	Restart      string
	RestartTip   string
	GitHubLink   string
	GitHubTip    string
	CheckUpdate  string
	CheckTip     string
	VersionLabel string
	Quit         string
	QuitTip      string
}

func GetTrayI18n(lang string, port int, ver string) TrayI18n {
	switch lang {
	case "zh":
		return TrayI18n{
			Tooltip:      fmt.Sprintf("Tim-Agent 在线终端 %s (端口: %d)", ver, port),
			TitleActive:  fmt.Sprintf("🟢 Tim-Agent %s 运行中 (:%d)", ver, port),
			OpenGUI:      "🌐 打开 Web 控制面板",
			OpenGUITip:   "在浏览器中打开配置与扫码页面",
			OpenTerm:     "📱 打开终端 (手机/PC)",
			OpenTermTip:  "在浏览器中打开终端界面",
			Restart:      "🔄 重启终端主会话",
			RestartTip:   "重启底层 Shell 会话并拉起Agent",
			GitHubLink:   "⭐ GitHub 官方开源主页",
			GitHubTip:    "在浏览器中访问 GitHub 仓库",
			CheckUpdate:  "🚀 检查新版本发布",
			CheckTip:     "在浏览器中查看 GitHub 最新 Release",
			VersionLabel: fmt.Sprintf("ℹ️ 当前版本: %s", ver),
			Quit:         "❌ 退出 Tim-Agent",
			QuitTip:      "停止服务并退出程序",
		}
	case "zh-TW":
		return TrayI18n{
			Tooltip:      fmt.Sprintf("Tim-Agent 線上終端 %s (連接埠: %d)", ver, port),
			TitleActive:  fmt.Sprintf("🟢 Tim-Agent %s 運行中 (:%d)", ver, port),
			OpenGUI:      "🌐 開啟 Web 控制面板",
			OpenGUITip:   "在瀏覽器中開啟設定與掃碼頁面",
			OpenTerm:     "📱 開啟終端機 (手機/PC)",
			OpenTermTip:  "在瀏覽器中開啟終端機介面",
			Restart:      "🔄 重啟終端機主工作階段",
			RestartTip:   "重啟底層 Shell 會話並拉起Agent",
			GitHubLink:   "⭐ GitHub 官方開源主頁",
			GitHubTip:    "在瀏覽器中造訪 GitHub 倉庫",
			CheckUpdate:  "🚀 檢查新版本發布",
			CheckTip:     "在瀏覽器中檢視 GitHub 最新 Release",
			VersionLabel: fmt.Sprintf("ℹ️ 目前版本: %s", ver),
			Quit:         "❌ 結束 Tim-Agent",
			QuitTip:      "停止服務並結束程式",
		}
	case "ja":
		return TrayI18n{
			Tooltip:      fmt.Sprintf("Tim-Agent オンライン端末 %s (ポート: %d)", ver, port),
			TitleActive:  fmt.Sprintf("🟢 Tim-Agent %s 実行中 (:%d)", ver, port),
			OpenGUI:      "🌐 Webコントロールセンターを開く",
			OpenGUITip:   "ブラウザで設定・QRコード画面を開く",
			OpenTerm:     "📱 端末を開く (スマホ/PC)",
			OpenTermTip:  "ブラウザで端末画面を開く",
			Restart:      "🔄 メインセッション再起動",
			RestartTip:   "シェルセッションを再起動してAgentを実行",
			GitHubLink:   "⭐ GitHub 公式リポジトリ",
			GitHubTip:    "ブラウザでGitHubを開く",
			CheckUpdate:  "🚀 最新バージョンを確認",
			CheckTip:     "ブラウザで最新Releaseを確認",
			VersionLabel: fmt.Sprintf("ℹ️ 現在のバージョン: %s", ver),
			Quit:         "❌ Tim-Agent を終了",
			QuitTip:      "サービスを停止して終了",
		}
	case "ko":
		return TrayI18n{
			Tooltip:      fmt.Sprintf("Tim-Agent 온라인 터미널 %s (포트: %d)", ver, port),
			TitleActive:  fmt.Sprintf("🟢 Tim-Agent %s 실행 중 (:%d)", ver, port),
			OpenGUI:      "🌐 웹 제어 센터 열기",
			OpenGUITip:   "브라우저에서 설정 및 QR 코드 페이지 열기",
			OpenTerm:     "📱 터미널 열기 (모바일/PC)",
			OpenTermTip:  "브라우저에서 터미널 화면 열기",
			Restart:      "🔄 메인 세션 재시작",
			RestartTip:   "쉘 세션을 재시작하고 Agent 실행",
			GitHubLink:   "⭐ GitHub 공식 저장소",
			GitHubTip:    "브라우저에서 GitHub 열기",
			CheckUpdate:  "🚀 최신 업데이트 확인",
			CheckTip:     "브라우저에서 최신 Release 확인",
			VersionLabel: fmt.Sprintf("ℹ️ 현재 버전: %s", ver),
			Quit:         "❌ Tim-Agent 종료",
			QuitTip:      "서비스를 중지하고 프로그램 종료",
		}
	case "de":
		return TrayI18n{
			Tooltip:      fmt.Sprintf("Tim-Agent Online-Terminal %s (Port: %d)", ver, port),
			TitleActive:  fmt.Sprintf("🟢 Tim-Agent %s Aktiv (:%d)", ver, port),
			OpenGUI:      "🌐 Web-Kontrollzentrum öffnen",
			OpenGUITip:   "Einstellungen und QR-Code im Browser öffnen",
			OpenTerm:     "📱 Terminal öffnen (Smartphone/PC)",
			OpenTermTip:  "Terminal-Oberfläche im Browser öffnen",
			Restart:      "🔄 Hauptsitzung neustarten",
			RestartTip:   "Shell-Sitzung und Agent neu starten",
			GitHubLink:   "⭐ GitHub Offizielles Repository",
			GitHubTip:    "GitHub im Browser öffnen",
			CheckUpdate:  "🚀 Nach Updates suchen",
			CheckTip:     "Neueste Releases im Browser prüfen",
			VersionLabel: fmt.Sprintf("ℹ️ Version: %s", ver),
			Quit:         "❌ Tim-Agent Beenden",
			QuitTip:      "Dienst stoppen und beenden",
		}
	default: // "en"
		return TrayI18n{
			Tooltip:      fmt.Sprintf("Tim-Agent Online Terminal %s (Port: %d)", ver, port),
			TitleActive:  fmt.Sprintf("🟢 Tim-Agent %s Running (:%d)", ver, port),
			OpenGUI:      "🌐 Open Control Center",
			OpenGUITip:   "Open settings and QR code in browser",
			OpenTerm:     "📱 Open Terminal (Mobile/PC)",
			OpenTermTip:  "Open web terminal interface",
			Restart:      "🔄 Restart Main Session",
			RestartTip:   "Restart shell session and launch Agent",
			GitHubLink:   "⭐ GitHub Official Repository",
			GitHubTip:    "Open GitHub repository in browser",
			CheckUpdate:  "🚀 Check for Updates",
			CheckTip:     "View latest releases on GitHub",
			VersionLabel: fmt.Sprintf("ℹ️ Version: %s", ver),
			Quit:         "❌ Quit Tim-Agent",
			QuitTip:      "Stop service and exit",
		}
	}
}
