//go:build !nocgo && (darwin || windows || linux)

package tray

import (
	"log"

	"fyne.io/systray"

	"github.com/tim-today/tim-agent/pkg/config"
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

var (
	currentTrayLang string
	updateTrayTexts func(string)
)

// UpdateTrayLanguage 动态更新托盘菜单显示的语言
func UpdateTrayLanguage(lang string) {
	if updateTrayTexts != nil {
		updateTrayTexts(lang)
	}
}

func RunTray(callbacks TrayCallbacks) {
	currentLang := callbacks.Language
	if currentLang == "" {
		currentLang = "en"
	}
	currentTrayLang = currentLang

	onReady := func() {
		iconBytes := GenerateTrayIcon()
		systray.SetIcon(iconBytes)
		systray.SetTitle("TimAgent")

		i18n := GetTrayI18n(currentTrayLang, callbacks.Port, config.AppVersion)
		systray.SetTooltip(i18n.Tooltip)

		mTitle := systray.AddMenuItem(i18n.TitleActive, "")
		mTitle.Disable()

		systray.AddSeparator()

		mOpenGUI := systray.AddMenuItem(i18n.OpenGUI, i18n.OpenGUITip)
		mOpenTerm := systray.AddMenuItem(i18n.OpenTerm, i18n.OpenTermTip)
		mRestart := systray.AddMenuItem(i18n.Restart, i18n.RestartTip)

		systray.AddSeparator()

		// 语言切换子菜单
		mLangMenu := systray.AddMenuItem("🌐 Language / 语言", "切换托盘界面语言")
		langItems := map[string]*systray.MenuItem{
			"en":    mLangMenu.AddSubMenuItem("English", "English"),
			"zh":    mLangMenu.AddSubMenuItem("简体中文", "简体中文"),
			"zh-TW": mLangMenu.AddSubMenuItem("繁體中文", "繁體中文"),
			"ja":    mLangMenu.AddSubMenuItem("日本語", "日本語"),
			"ko":    mLangMenu.AddSubMenuItem("한국어", "한국어"),
			"de":    mLangMenu.AddSubMenuItem("Deutsch", "Deutsch"),
		}

		refreshLangChecks := func(activeLang string) {
			for l, item := range langItems {
				if l == activeLang {
					item.Check()
				} else {
					item.Uncheck()
				}
			}
		}
		refreshLangChecks(currentTrayLang)

		systray.AddSeparator()

		// 官方链接与版本检查
		mGitHub := systray.AddMenuItem(i18n.GitHubLink, i18n.GitHubTip)
		mCheckUpdate := systray.AddMenuItem(i18n.CheckUpdate, i18n.CheckTip)
		mVerInfo := systray.AddMenuItem(i18n.VersionLabel, "")
		mVerInfo.Disable()

		updateTrayTexts = func(lang string) {
			currentTrayLang = lang
			t := GetTrayI18n(lang, callbacks.Port, config.AppVersion)
			systray.SetTooltip(t.Tooltip)
			mTitle.SetTitle(t.TitleActive)
			mOpenGUI.SetTitle(t.OpenGUI)
			mOpenGUI.SetTooltip(t.OpenGUITip)
			mOpenTerm.SetTitle(t.OpenTerm)
			mOpenTerm.SetTooltip(t.OpenTermTip)
			mRestart.SetTitle(t.Restart)
			mRestart.SetTooltip(t.RestartTip)
			mGitHub.SetTitle(t.GitHubLink)
			mGitHub.SetTooltip(t.GitHubTip)
			mCheckUpdate.SetTitle(t.CheckUpdate)
			mCheckUpdate.SetTooltip(t.CheckTip)
			mVerInfo.SetTitle(t.VersionLabel)
			refreshLangChecks(lang)
		}

		systray.AddSeparator()

		mQuit := systray.AddMenuItem(i18n.Quit, i18n.QuitTip)

		for l, item := range langItems {
			targetLang := l
			menuItem := item
			go func() {
				for range menuItem.ClickedCh {
					updateTrayTexts(targetLang)
					if callbacks.OnLanguageChange != nil {
						callbacks.OnLanguageChange(targetLang)
					}
				}
			}()
		}

		go func() {
			for {
				select {
				case <-mOpenGUI.ClickedCh:
					if callbacks.OnOpenGUI != nil {
						callbacks.OnOpenGUI()
					}
				case <-mOpenTerm.ClickedCh:
					if callbacks.OnOpenTerm != nil {
						callbacks.OnOpenTerm()
					}
				case <-mRestart.ClickedCh:
					if callbacks.OnRestartPty != nil {
						callbacks.OnRestartPty()
					}
				case <-mGitHub.ClickedCh:
					_ = OpenURL(config.GitHubRepoURL)
				case <-mCheckUpdate.ClickedCh:
					_ = OpenURL(config.GitHubReleases)
				case <-mQuit.ClickedCh:
					systray.Quit()
					if callbacks.OnExit != nil {
						callbacks.OnExit()
					}
					return
				}
			}
		}()
	}

	onExit := func() {
		log.Println("[TimAgent] 托盘已退出")
	}

	systray.Run(onReady, onExit)
}
