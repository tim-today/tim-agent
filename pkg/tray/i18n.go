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
