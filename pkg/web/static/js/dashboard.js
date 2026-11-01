// 控制面板前端交互
(function () {
  'use strict';

  // 多语言 i18n 字典系统 (默认 English)
  const I18N = {
    en: {
      app_title: "Tim-Agent Control Center",
      status_running: "Active",
      btn_enter_terminal: "Open Terminal ↗",
      card1_title: "📱 Mobile Access via QR Code (Tailscale / LAN)",
      card1_subtitle: "Scan with phone camera or browser to connect directly",
      qr_loading: "Generating...",
      label_network: "Select Network Interface:",
      btn_copy: "Copy Link",
      btn_copied: "Copied!",
      card1_tip: "💡 With Tailscale, your phone and computer can connect directly anywhere over your private mesh VPN with zero port forwarding!",
      card2_title: "🖥️ Terminal Session (Persistent)",
      btn_refresh: "Refresh",
      status_running_badge: "Running",
      sess_main_title: "Main Session (default)",
      btn_open_terminal: "Open Terminal",
      btn_restart_agent: "🔄 Restart & Launch Agent",
      card2_tip: "🔒 Session Keep-Alive Active: Background shell process continues even if phone locks or browser closes!",
      card3_title: "⚙️ Service & Security Configuration",
      label_port: "Shared Terminal Port:",
      hint_port: "Restart required after changing port",
      label_password: "Access Password (Terminal Security):",
      btn_regen_pwd: "🎲 Random",
      hint_password: "Saved permanently in mobile/PC browser after first entry",
      label_workdir: "Bound Working Directory (WorkDir):",
      hint_workdir: "Security isolation: terminal & agent run in this directory",
      label_agent: "CLI Agent Preset:",
      label_auto_approve: "⚡ Auto-Approve / Bypass Sandbox (Default off)",
      hint_agent: "Auto-restarts session and launches selected Agent upon saving",
      label_shell: "Shell Path:",
      label_theme: "Default Terminal Theme:",
      label_language: "UI Language (界面语言):",
      label_keep_alive: "Session Keep-Alive:",
      hint_keep_alive: "Keep tasks running after disconnection",
      btn_save_config: "Save Settings",
      saving_text: "Saving...",
      saved_text: "Settings saved!",
      saved_restarted_prefix: "Settings saved & launched Agent [",
      prompt_copy: "Please copy link manually:",
      confirm_regen: "Generate a new random access password?",
      regen_success: "New password generated & active: ",
      auto_approve_yes: "【Auto-Approve Enabled】",
      auto_approve_no: "【Standard Safe Mode】",
      restart_success: "Terminal session restarted and Agent launched!",
      restart_fail: "Failed to restart: ",
      net_error: "Network error",
      btn_check_update: "🚀 Update",
      checking_update: "Checking for updates...",
      up_to_date: "✓ Up to date",
      update_available: "⭐ New version available: ",
      check_failed: "Update check failed",
      click_to_update: "Click to download update",
      footer_releases: "Releases & Changelog",
      label_remote_access: "Remote Management (Web Dashboard):",
      hint_remote_access: "Default off (127.0.0.1/localhost only). When enabled, external IPs can access dashboard with password.",
      badge_remote_off: "Localhost Only",
      badge_remote_on: "Remote Allowed",
      card_projects_title: "📂 Project Management (Multi-Session)",
      card_projects_subtitle: "Configure multiple projects with independent WorkDir & CLI Agent. Remote access enters the most recently operated session by default; background sessions keep running when switching. (Same WorkDir cannot be shared by different projects)",
      btn_add_project: "+ Add Project",
      modal_proj_add_title: "📂 Add New Project",
      modal_proj_edit_title: "✏️ Edit Project",
      label_proj_name: "Project Name:",
      label_proj_workdir: "Working Directory (WorkDir):",
      hint_proj_workdir: "⚠️ Each project must have a unique directory. Same directory cannot belong to different projects.",
      label_proj_agent: "CLI Agent Preset:",
      label_proj_auto_approve: "⚡ Auto-Approve / Bypass Sandbox",
      label_proj_set_active: "Set as current active project (default for remote connection)",
      btn_cancel: "Cancel",
      btn_save_project: "Save Project",
      badge_active_recent: "⭐ Recent Active",
      badge_standby: "Standby (Not Started)",
      btn_switch_open: "🚀 Open Terminal",
      btn_set_active: "▶️ Activate / Start",
      btn_restart_proj: "🔄 Restart",
      btn_edit_proj: "✏️ Edit",
      btn_del_proj: "🗑️ Delete",
      confirm_del_proj: "Delete project ",
      dup_workdir_err: "Working directory already belongs to another project: ",
    },
    zh: {
      app_title: "Tim-Agent 控制面板",
      status_running: "常驻运行中",
      btn_enter_terminal: "进入终端 ↗",
      btn_check_update: "🚀 检查更新",
      checking_update: "正在检查最新版本...",
      up_to_date: "✓ 当前已是最新版本",
      update_available: "⭐ 发现新版本: ",
      check_failed: "检查更新失败",
      click_to_update: "点击前往下载新版本",
      footer_releases: "版本发布与更新日志",
      card1_title: "📱 手机扫码连接 (Tailscale / 局域网)",
      card1_subtitle: "手机使用相机或浏览器直接扫码免密接入",
      qr_loading: "生成中...",
      label_network: "选择连接网络:",
      btn_copy: "复制链接",
      btn_copied: "已复制!",
      card1_tip: "💡 配合 Tailscale，手机和电脑处在同一虚拟内网即可随时随地直连，无需公网IP和端口映射！",
      card2_title: "🖥️ 终端会话状态 (永不中断)",
      btn_refresh: "刷新",
      status_running_badge: "运行中",
      sess_main_title: "主会话 (default)",
      btn_open_terminal: "打开终端",
      btn_restart_agent: "🔄 重启并拉起Agent",
      card2_tip: "🔒 会话保持开启中：即使手机锁屏、浏览器关闭、网络切断，后台 Shell 进程依然持续运行！",
      card3_title: "⚙️ 服务与安全配置",
      label_port: "共享服务端口:",
      hint_port: "修改端口需重启服务生效",
      label_password: "访问密码 (保护终端安全):",
      btn_regen_pwd: "🎲 重新随机",
      hint_password: "Web访问输入一次自动保存，扫码亦可一键免密直连",
      label_workdir: "当前活动项目工作目录 (WorkDir):",
      hint_workdir: "安全隔离：终端与 Agent 命令均在此目录内运行",
      label_agent: "当前活动项目 Agent 代理:",
      label_auto_approve: "⚡ 自动确认 / 关闭沙箱 (默认不选)",
      hint_agent: "切换后点击“保存配置”，系统将自动重启当前项目会话并拉起所选 Agent",
      label_shell: "执行 Shell 路径:",
      label_theme: "默认终端配色方案:",
      label_language: "界面语言 (Language):",
      label_keep_alive: "会话持久保活 (Keep-Alive):",
      hint_keep_alive: "断开连接后保持任务继续执行",
      btn_save_config: "保存配置",
      saving_text: "保存中...",
      saved_text: "配置已保存！",
      saved_restarted_prefix: "配置已保存，已自动切换并拉起 [",
      prompt_copy: "请手动复制链接:",
      confirm_regen: "确定重新生成一个随机访问密码吗？",
      regen_success: "新密码已生成并生效: ",
      auto_approve_yes: "【开启自动确认/关闭沙箱】",
      auto_approve_no: "【普通安全确认】",
      restart_success: "终端已重启，并已拉起指定 Agent！",
      restart_fail: "重启失败: ",
      net_error: "网络错误",
      label_remote_access: "远程管理链接 (控制面板):",
      hint_remote_access: "默认关闭 (仅限本机 127.0.0.1/localhost 访问)。开启后允许外部 IP 输入密码访问控制面板。",
      badge_remote_off: "仅限本机",
      badge_remote_on: "允许外部IP",
      card_projects_title: "📂 项目管理 (多会话多项目切换)",
      card_projects_subtitle: "可配置多个独立项目（各自拥有项目名、WorkDir、Agent、是否自动确认）。远程链接默认进入最近操作的会话，切换项目后之前会话保持在后台运行。（注意：不能同一个目录属于不同项目）",
      btn_add_project: "+ 新增项目",
      modal_proj_add_title: "📂 新增项目配置",
      modal_proj_edit_title: "✏️ 编辑项目配置",
      label_proj_name: "项目名称 (Project Name):",
      label_proj_workdir: "项目工作目录 (WorkDir):",
      hint_proj_workdir: "⚠️ 限制：不能将同一个目录分配给不同项目",
      label_proj_agent: "命令行 Agent 代理:",
      label_proj_auto_approve: "⚡ 自动确认 / 关闭沙箱 (默认不选)",
      label_proj_set_active: "设为当前最近活动项目 (远程链接默认进入此会话)",
      btn_cancel: "取消",
      btn_save_project: "保存项目",
      badge_active_recent: "⭐ 最近活动 (默认进入)",
      badge_standby: "未启动 (待机)",
      btn_switch_open: "🚀 进入终端",
      btn_set_active: "▶️ 启动 / 设为当前",
      btn_restart_proj: "🔄 重启",
      btn_edit_proj: "✏️ 编辑",
      btn_del_proj: "🗑️ 删除",
      confirm_del_proj: "确定要删除项目 ",
      dup_workdir_err: "工作目录不能重复，该目录已属于项目：",
    },
    'zh-TW': {
      app_title: "Tim-Agent 控制面板",
      status_running: "常駐運行中",
      btn_enter_terminal: "進入終端 ↗",
      card1_title: "📱 手機掃碼連接 (Tailscale / 區域網路)",
      card1_subtitle: "手機使用相機或瀏覽器直接掃碼免密接入",
      qr_loading: "生成中...",
      label_network: "選擇連接網路:",
      btn_copy: "複製連結",
      btn_copied: "已複製!",
      card1_tip: "💡 配合 Tailscale，手機與電腦處於同一虛擬內網即可隨時隨地直連，無需公網IP！",
      card2_title: "🖥️ 終端會話狀態 (永不中斷)",
      btn_refresh: "重新整理",
      status_running_badge: "運行中",
      sess_main_title: "主會話 (default)",
      btn_open_terminal: "開啟終端",
      btn_restart_agent: "🔄 重啟並拉起Agent",
      card2_tip: "🔒 會話保持開啟中：即使手機鎖定、瀏覽器關閉，後台 Shell 依舊持續運行！",
      card3_title: "⚙️ 服務與安全設定",
      label_port: "共享服務埠號:",
      hint_port: "修改埠號需重啟生效",
      label_password: "存取密碼 (保護終端安全):",
      btn_regen_pwd: "🎲 重新隨機",
      hint_password: "Web存取輸入一次自動保存，掃碼亦可一鍵免密直連",
      label_workdir: "安全綁定工作目錄 (WorkDir):",
      hint_workdir: "安全隔離：終端與 Agent 命令均在此目錄內運行",
      label_agent: "選擇命令列 Agent 代理:",
      label_auto_approve: "⚡ 自動確認 / 關閉沙箱 (預設不選)",
      hint_agent: "切換後點擊“保存設定”，系統將自動重啟主會話並拉起所選 Agent",
      label_shell: "執行 Shell 路徑:",
      label_theme: "預設終端配色主題:",
      label_language: "介面語言 (Language):",
      label_keep_alive: "會話持久保活 (Keep-Alive):",
      hint_keep_alive: "斷開連線後保持任務繼續執行",
      btn_save_config: "儲存設定",
      saving_text: "儲存中...",
      saved_text: "設定已儲存！",
      saved_restarted_prefix: "設定已儲存，已自動切換並拉起 [",
      prompt_copy: "請手動複製連結:",
      confirm_regen: "確定重新生成一個隨機存取密碼嗎？",
      regen_success: "新密碼已生成並生效: ",
      auto_approve_yes: "【開啟自動確認/關閉沙箱】",
      auto_approve_no: "【標準安全模式】",
      restart_success: "終端已重啟，並已拉起指定 Agent！",
      restart_fail: "重啟失敗: ",
      net_error: "網路錯誤",
      btn_check_update: "🚀 檢查更新",
      checking_update: "正在檢查最新版本...",
      up_to_date: "✓ 目前已是最新版本",
      update_available: "⭐ 發現新版本: ",
      check_failed: "檢查更新失敗",
      click_to_update: "點擊前往下載新版本",
      footer_releases: "版本發布與更新日誌",
      label_remote_access: "遠端管理連結 (控制面板):",
      hint_remote_access: "預設關閉 (僅限本機 127.0.0.1/localhost 存取)。開啟後允許外部 IP 輸入密碼存取控制面板。",
      badge_remote_off: "僅限本機",
      badge_remote_on: "允許外部IP",
    },
    ja: {
      app_title: "Tim-Agent コントロールセンター",
      status_running: "実行中",
      btn_enter_terminal: "端末を開く ↗",
      btn_check_update: "🚀 更新確認",
      checking_update: "最新バージョンを確認中...",
      up_to_date: "✓ 最新バージョンです",
      update_available: "⭐ 新バージョンが利用可能: ",
      check_failed: "更新確認に失敗しました",
      click_to_update: "クリックしてダウンロード",
      footer_releases: "リリースと更新履歴",
      card1_title: "📱 QRコードでスマホ接続 (Tailscale / LAN)",
      card1_subtitle: "スマホのカメラまたはブラウザでQRをスキャンして直接接続",
      qr_loading: "生成中...",
      label_network: "接続ネットワークを選択:",
      btn_copy: "リンクをコピー",
      btn_copied: "コピー完了!",
      card1_tip: "💡 Tailscale を使用すると、ポート転送なしでどこからでも直接接続できます！",
      card2_title: "🖥️ 端末セッション状態 (常駐)",
      btn_refresh: "更新",
      status_running_badge: "実行中",
      sess_main_title: "メインセッション (default)",
      btn_open_terminal: "端末を開く",
      btn_restart_agent: "🔄 再起動してAgent起動",
      card2_tip: "🔒 セッション永続化有効: 画面ロックやブラウザ終了後もバックグラウンドで継続します！",
      card3_title: "⚙️ サービス＆セキュリティ設定",
      label_port: "共有端末ポート:",
      hint_port: "ポート変更後は再起動が必要",
      label_password: "アクセスパスワード (端末保護):",
      btn_regen_pwd: "🎲 ランダム生成",
      hint_password: "初回入力後ブラウザに自動保存されます",
      label_workdir: "バインド作業ディレクトリ (WorkDir):",
      hint_workdir: "セキュリティ分離: 端末とAgentはこのディレクトリで実行されます",
      label_agent: "CLI Agent プリセット:",
      label_auto_approve: "⚡ 自動承認 / サンドボックス無効化 (標準OFF)",
      hint_agent: "保存時にセッションを再起動し選択したAgentを自動起動します",
      label_shell: "シェルパス:",
      label_theme: "デフォルト端末テーマ:",
      label_language: "UI 言語 (Language):",
      label_keep_alive: "セッション常駐保持:",
      hint_keep_alive: "切断後もバックグラウンドでタスクを継続実行",
      btn_save_config: "設定を保存",
      saving_text: "保存中...",
      saved_text: "設定を保存しました！",
      saved_restarted_prefix: "設定を保存しAgentを起動しました [",
      prompt_copy: "リンクを手動でコピーしてください:",
      confirm_regen: "新しいランダムパスワードを生成しますか？",
      regen_success: "新しいパスワードが適用されました: ",
      auto_approve_yes: "【自動承認有効】",
      auto_approve_no: "【標準セーフモード】",
      restart_success: "端末セッションが再起動しAgentが起動しました！",
      restart_fail: "再起動に失敗しました: ",
      net_error: "ネットワークエラー",
      label_remote_access: "リモート管理 (ダッシュボード):",
      hint_remote_access: "デフォルト無効 (127.0.0.1/localhost のみ)。有効にすると外部IPからパスワード認証で管理可能。",
      badge_remote_off: "ローカル専用",
      badge_remote_on: "外部許可",
    },
    ko: {
      app_title: "Tim-Agent 제어 센터",
      status_running: "실행 중",
      btn_enter_terminal: "터미널 열기 ↗",
      btn_check_update: "🚀 업데이트 확인",
      checking_update: "최신 버전 확인 중...",
      up_to_date: "✓ 현재 최신 버전입니다",
      update_available: "⭐ 새 버전 사용 가능: ",
      check_failed: "업데이트 확인 실패",
      click_to_update: "클릭하여 다운로드",
      footer_releases: "릴리즈 및 변경 로그",
      card1_title: "📱 QR 코드로 모바일 접속 (Tailscale / LAN)",
      card1_subtitle: "모바일 카메라나 브라우저로 스캔하여 즉시 연결",
      qr_loading: "생성 중...",
      label_network: "연결 네트워크 선택:",
      btn_copy: "링크 복사",
      btn_copied: "복사됨!",
      card1_tip: "💡 Tailscale과 함께라면 포트포워딩 없이 어디서나 안전하게 직접 연결됩니다!",
      card2_title: "🖥️ 터미널 세션 상태 (영구 유지)",
      btn_refresh: "새로고침",
      status_running_badge: "실행 중",
      sess_main_title: "메인 세션 (default)",
      btn_open_terminal: "터미널 열기",
      btn_restart_agent: "🔄 재시작 및 Agent 실행",
      card2_tip: "🔒 세션 유지 활성화: 모바일 화면 잠금이나 브라우저 종료 시에도 계속 실행됩니다!",
      card3_title: "⚙️ 서비스 및 보안 구성",
      label_port: "공유 터미널 포트:",
      hint_port: "포트 변경 후 재시작 필요",
      label_password: "접속 비밀번호 (터미널 보안):",
      btn_regen_pwd: "🎲 무작위 생성",
      hint_password: "최초 입력 시 브라우저에 영구 저장됩니다",
      label_workdir: "바인딩 작업 디렉토리 (WorkDir):",
      hint_workdir: "보안 격리: 터미널과 Agent는 이 디렉토리에서 실행됩니다",
      label_agent: "CLI Agent 프리셋:",
      label_auto_approve: "⚡ 자동 확인 / 샌드박스 비활성화 (기본 OFF)",
      hint_agent: "저장 시 세션을 재시작하고 선택한 Agent를 자동 실행합니다",
      label_shell: "쉘 경로:",
      label_theme: "기본 터미널 테마:",
      label_language: "UI 언어 (Language):",
      label_keep_alive: "세션 유지 (Keep-Alive):",
      hint_keep_alive: "연결이 끊겨도 백그라운드에서 작업 계속 유지",
      btn_save_config: "설정 저장",
      saving_text: "저장 중...",
      saved_text: "설정이 저장되었습니다!",
      saved_restarted_prefix: "설정이 저장되고 Agent가 실행되었습니다 [",
      prompt_copy: "링크를 수동으로 복사하세요:",
      confirm_regen: "새로운 무작위 비밀번호를 생성하시겠습니까?",
      regen_success: "새 비밀번호가 생성되었습니다: ",
      auto_approve_yes: "【자동 확인 활성화】",
      auto_approve_no: "【표준 안전 모드】",
      restart_success: "터미널 세션이 재시작되고 Agent가 실행되었습니다!",
      restart_fail: "재시작 실패: ",
      net_error: "네트워크 오류",
      label_remote_access: "원격 관리 (대시보드):",
      hint_remote_access: "기본 비활성화 (127.0.0.1/localhost 전용). 활성화 시 외부 IP에서 비밀번호로 접속 가능.",
      badge_remote_off: "로컬 전용",
      badge_remote_on: "외부 허용",
    },
    de: {
      app_title: "Tim-Agent Kontrollzentrum",
      status_running: "Aktiv",
      btn_enter_terminal: "Terminal öffnen ↗",
      btn_check_update: "🚀 Update prüfen",
      checking_update: "Suche nach Updates...",
      up_to_date: "✓ Auf dem neuesten Stand",
      update_available: "⭐ Neue Version verfügbar: ",
      check_failed: "Update-Prüfung fehlgeschlagen",
      click_to_update: "Hier klicken zum Herunterladen",
      footer_releases: "Releases & Changelog",
      card1_title: "📱 Mobiler Zugriff per QR-Code (Tailscale / LAN)",
      card1_subtitle: "Mit Smartphone-Kamera oder Browser scannen für Direktzugriff",
      qr_loading: "Wird generiert...",
      label_network: "Netzwerkschnittstelle wählen:",
      btn_copy: "Link kopieren",
      btn_copied: "Kopiert!",
      card1_tip: "💡 Mit Tailscale überall ohne Portweiterleitung sicher verbinden!",
      card2_title: "🖥️ Terminal-Sitzung (Dauerhaft)",
      btn_refresh: "Aktualisieren",
      status_running_badge: "Aktiv",
      sess_main_title: "Hauptsitzung (default)",
      btn_open_terminal: "Terminal öffnen",
      btn_restart_agent: "🔄 Neustart & Agent ausführen",
      card2_tip: "🔒 Sitzung bleibt aktiv: Hintergrundprozess läuft auch bei Bildschirmsperre weiter!",
      card3_title: "⚙️ Dienst- & Sicherheitskonfiguration",
      label_port: "Gemeinsamer Terminal-Port:",
      hint_port: "Neustart nach Portänderung erforderlich",
      label_password: "Zugriffskennwort (Terminalsicherheit):",
      btn_regen_pwd: "🎲 Zufall",
      hint_password: "Wird nach Ersteingabe dauerhaft im Browser gespeichert",
      label_workdir: "Arbeitsverzeichnis (WorkDir):",
      hint_workdir: "Sicherheitsisolation: Terminal & Agent laufen in diesem Verzeichnis",
      label_agent: "CLI Agent Preset:",
      label_auto_approve: "⚡ Automatische Bestätigung (Standard AUS)",
      hint_agent: "Startet die Sitzung neu und führt gewählten Agent aus",
      label_shell: "Shell-Pfad:",
      label_theme: "Standard Terminal-Farbschema:",
      label_language: "Sprache (Language):",
      label_keep_alive: "Sitzung aufrechterhalten:",
      hint_keep_alive: "Aufgaben nach Trennung weiter ausführen",
      btn_save_config: "Einstellungen speichern",
      saving_text: "Speichern...",
      saved_text: "Einstellungen gespeichert!",
      saved_restarted_prefix: "Einstellungen gespeichert & Agent gestartet [",
      prompt_copy: "Link bitte manuell kopieren:",
      confirm_regen: "Neues Zufallskennwort generieren?",
      regen_success: "Neues Kennwort aktiv: ",
      auto_approve_yes: "【Auto-Approve aktiv】",
      auto_approve_no: "【Sicherer Standardmodus】",
      restart_success: "Terminal neu gestartet und Agent ausgeführt!",
      restart_fail: "Neustart fehlgeschlagen: ",
      net_error: "Netzwerkfehler",
      label_remote_access: "Fernverwaltung (Dashboard):",
      hint_remote_access: "Standardmäßig aus (nur 127.0.0.1/localhost). Wenn aktiviert, können externe IPs mit Passwort zugreifen.",
      badge_remote_off: "Nur Lokal",
      badge_remote_on: "Extern Erlaubt",
    }
  };

  let currentLang = localStorage.getItem('tim_lang') || 'en';

  function applyLanguage(lang) {
    if (!I18N[lang]) lang = 'en';
    currentLang = lang;
    localStorage.setItem('tim_lang', lang);
    document.documentElement.lang = lang;

    document.querySelectorAll('[data-i18n]').forEach(el => {
      const key = el.getAttribute('data-i18n');
      if (I18N[lang] && I18N[lang][key]) {
        el.textContent = I18N[lang][key];
      }
    });

    const cfgLangEl = document.getElementById('cfgLanguage');
    if (cfgLangEl) cfgLangEl.value = lang;
  }

  let currentStatus = null;

  const networkSelect = document.getElementById('networkSelect');
  const qrImage = document.getElementById('qrImage');
  const qrPlaceholder = document.getElementById('qrPlaceholder');
  const currentUrlText = document.getElementById('currentUrlText');
  const btnCopyUrl = document.getElementById('btnCopyUrl');

  const sessPid = document.getElementById('sessPid');
  const sessClients = document.getElementById('sessClients');
  const btnRestartPty = document.getElementById('btnRestartPty');
  const btnRefreshSession = document.getElementById('btnRefreshSession');

  const configForm = document.getElementById('configForm');
  const cfgPort = document.getElementById('cfgPort');
  const cfgPassword = document.getElementById('cfgPassword');
  const cfgShell = document.getElementById('cfgShell');
  const cfgKeepAlive = document.getElementById('cfgKeepAlive');
  const cfgLanguage = document.getElementById('cfgLanguage');
  const saveStatus = document.getElementById('saveStatus');

  async function fetchStatus() {
    try {
      const res = await fetch('/api/status');
      const data = await res.json();
      currentStatus = data;
      if (data.version) {
        if (appVersionBadge) appVersionBadge.textContent = data.version;
        if (footerVersionText) footerVersionText.textContent = data.version;
      }
      renderNetworks(data.networks, data.auth_token);
      renderConfig(data.config);
      renderSession(data.session);
      renderProjectsList(data.projects_status || (data.config && data.config.projects) || [], data.active_project_id);
    } catch (err) {
      console.error('获取状态失败', err);
    }
  }

  function renderNetworks(networks, token) {
    networkSelect.innerHTML = '';
    let selectedIndex = 0;

    networks.forEach((net, idx) => {
      const opt = document.createElement('option');
      opt.value = net.url + '/terminal' + (token ? '?token=' + encodeURIComponent(token) : '');
      opt.textContent = `${net.label} (${net.ip}) [${net.interface}]`;
      // 优先默认选中 Tailscale，其次选中局域网 IP
      if (net.is_vpn) {
        selectedIndex = idx;
      } else if (selectedIndex === 0 && net.is_private) {
        selectedIndex = idx;
      }
      networkSelect.appendChild(opt);
    });

    if (networks.length > 0) {
      networkSelect.selectedIndex = selectedIndex;
      updateQRCode();
    }
  }

  function updateQRCode() {
    const targetUrl = networkSelect.value;
    if (!targetUrl) return;

    currentUrlText.textContent = targetUrl;
    qrPlaceholder.style.display = 'block';
    qrImage.style.display = 'none';

    const qrSrc = '/api/qrcode?url=' + encodeURIComponent(targetUrl);
    qrImage.src = qrSrc;
    qrImage.onload = function () {
      qrPlaceholder.style.display = 'none';
      qrImage.style.display = 'block';
    };
  }

  networkSelect.addEventListener('change', updateQRCode);

  btnCopyUrl.addEventListener('click', async function () {
    const url = currentUrlText.textContent;
    try {
      await navigator.clipboard.writeText(url);
      btnCopyUrl.textContent = I18N[currentLang].btn_copied || 'Copied!';
      setTimeout(() => { btnCopyUrl.textContent = I18N[currentLang].btn_copy || 'Copy Link'; }, 1500);
    } catch (e) {
      prompt(I18N[currentLang].prompt_copy || 'Please copy link manually:', url);
    }
  });

  const cfgWorkDir = document.getElementById('cfgWorkDir');
  const cfgAgent = document.getElementById('cfgAgent');
  const cfgAutoApprove = document.getElementById('cfgAutoApprove');
  const cfgRemoteAccess = document.getElementById('cfgRemoteAccess');
  const remoteAccessBadge = document.getElementById('remoteAccessBadge');
  const cfgTheme = document.getElementById('cfgTheme');
  const sessAgent = document.getElementById('sessAgent');
  const sessWorkDir = document.getElementById('sessWorkDir');

  function updateRemoteAccessBadge(enabled) {
    if (!remoteAccessBadge) return;
    if (enabled) {
      remoteAccessBadge.textContent = I18N[currentLang].badge_remote_on || '允许外部IP';
      remoteAccessBadge.className = 'remote-badge remote-enabled';
    } else {
      remoteAccessBadge.textContent = I18N[currentLang].badge_remote_off || '仅限本机';
      remoteAccessBadge.className = 'remote-badge remote-disabled';
    }
  }

  if (cfgRemoteAccess) {
    cfgRemoteAccess.addEventListener('change', async function () {
      const enabled = this.checked;
      updateRemoteAccessBadge(enabled);
      if (currentStatus && currentStatus.config) {
        const cfg = currentStatus.config;
        try {
          await fetch('/api/config', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({
              port: cfg.port,
              password: cfg.password,
              shell: cfg.shell,
              work_dir: cfg.work_dir,
              selected_agent: cfg.selected_agent,
              theme: cfg.theme,
              language: cfg.language || currentLang,
              auto_approve: cfg.auto_approve,
              keep_alive: cfg.keep_alive,
              remote_access: enabled
            })
          });
          currentStatus.config.remote_access = enabled;
        } catch (e) {
          console.error('更新远程管理设置失败', e);
        }
      }
    });
  }

  function onLanguageChange(newLang) {
    applyLanguage(newLang);
    updateRemoteAccessBadge(cfgRemoteAccess ? cfgRemoteAccess.checked : false);
    // 异步快速保存语言偏好至服务器
    if (currentStatus && currentStatus.config) {
      const cfg = currentStatus.config;
      fetch('/api/config', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          port: cfg.port,
          password: cfg.password,
          shell: cfg.shell,
          work_dir: cfg.work_dir,
          selected_agent: cfg.selected_agent,
          theme: cfg.theme,
          language: newLang,
          auto_approve: cfg.auto_approve,
          keep_alive: cfg.keep_alive,
          remote_access: cfg.remote_access
        })
      }).catch(() => {});
    }
  }

  if (cfgLanguage) {
    cfgLanguage.addEventListener('change', function () {
      onLanguageChange(this.value);
    });
  }

  function renderConfig(cfg) {
    if (!cfg) return;
    if (document.activeElement !== cfgPort) cfgPort.value = cfg.port;
    if (document.activeElement !== cfgPassword) cfgPassword.value = cfg.password || '';
    if (document.activeElement !== cfgShell) cfgShell.value = cfg.shell || '';
    if (cfgWorkDir && document.activeElement !== cfgWorkDir) cfgWorkDir.value = cfg.work_dir || '';
    if (document.activeElement !== cfgAutoApprove) cfgAutoApprove.checked = !!cfg.auto_approve;
    if (cfgTheme && cfg.theme && document.activeElement !== cfgTheme) {
      cfgTheme.value = cfg.theme;
    }
    const lang = cfg.language || currentLang || 'en';
    applyLanguage(lang);
    if (document.activeElement !== cfgKeepAlive) cfgKeepAlive.checked = cfg.keep_alive;
    if (cfgRemoteAccess) {
      cfgRemoteAccess.checked = !!cfg.remote_access;
      updateRemoteAccessBadge(!!cfg.remote_access);
    }

    // 填充 Agent 选项（保留项目弹窗内正在编辑的 Agent 选择值，防止定时轮询冲掉用户选择）
    if (cfg.agents && cfg.agents.length > 0) {
      if (document.activeElement !== cfgAgent) {
        cfgAgent.innerHTML = '';
        cfg.agents.forEach(a => {
          const opt = document.createElement('option');
          opt.value = a.id;
          opt.textContent = `${a.name} (${a.command || 'shell'}) - ${a.description}`;
          if (a.id === cfg.selected_agent) {
            opt.selected = true;
          }
          cfgAgent.appendChild(opt);
        });
      }

      const isProjModalOpen = projectModal && projectModal.style.display !== 'none';
      if (projAgentSelect && (!isProjModalOpen || projAgentSelect.options.length === 0)) {
        const prevProjAgent = projAgentSelect.value;
        projAgentSelect.innerHTML = '';
        cfg.agents.forEach(a => {
          const opt2 = document.createElement('option');
          opt2.value = a.id;
          opt2.textContent = `${a.name} (${a.command || 'shell'}) - ${a.description}`;
          projAgentSelect.appendChild(opt2);
        });
        if (prevProjAgent) {
          projAgentSelect.value = prevProjAgent;
        }
      }
    }
  }

  const sessName = document.getElementById('sessName');

  function renderSession(sess) {
    if (!sess) return;
    if (sessName && sess.name) {
      sessName.textContent = `${sess.name} (${sess.id || 'default'})`;
    }
    sessPid.textContent = `PID: ${sess.pid || '--'}`;
    const clientsPrefix = I18N[currentLang].clients_label || 'Clients:';
    sessClients.textContent = `${clientsPrefix} ${sess.clients_count || 0}`;
    if (sess.active_agent) {
      const isAuto = sess.auto_approve !== undefined ? sess.auto_approve : (currentStatus && currentStatus.config && currentStatus.config.auto_approve);
      const approveBadge = isAuto ? ' [⚡Auto-Approve]' : '';
      sessAgent.textContent = `Agent: ${sess.active_agent}${approveBadge}`;
    }
    if (sess.work_dir) {
      const workDirPrefix = I18N[currentLang].workdir_label || 'WorkDir:';
      sessWorkDir.textContent = `📁 ${workDirPrefix} ${sess.work_dir}`;
    }
  }

  // 多项目管理渲染与操作
  const projectsListContainer = document.getElementById('projectsListContainer');
  const btnAddProject = document.getElementById('btnAddProject');
  const projectModal = document.getElementById('projectModal');
  const projectModalTitle = document.getElementById('projectModalTitle');
  const btnCloseProjectModal = document.getElementById('btnCloseProjectModal');
  const btnCancelProjectModal = document.getElementById('btnCancelProjectModal');
  const projectForm = document.getElementById('projectForm');
  const projIdInput = document.getElementById('projIdInput');
  const projNameInput = document.getElementById('projNameInput');
  const projWorkDirInput = document.getElementById('projWorkDirInput');
  const projAgentSelect = document.getElementById('projAgentSelect');
  const projAutoApproveCheck = document.getElementById('projAutoApproveCheck');
  const projSetActiveCheck = document.getElementById('projSetActiveCheck');
  const projErrorMsg = document.getElementById('projErrorMsg');
  const tabBtnBasic = document.getElementById('tabBtnBasic');
  const tabBtnDeploy = document.getElementById('tabBtnDeploy');
  const tabPaneBasic = document.getElementById('tabPaneBasic');
  const tabPaneDeploy = document.getElementById('tabPaneDeploy');
  const btnAddDeployApp = document.getElementById('btnAddDeployApp');
  const deployAppsContainer = document.getElementById('deployAppsContainer');

  // 当前弹窗内正在编辑的多应用多环境部署列表与当前正在编辑的应用索引 (-1 表示列表视图)
  let editingDeployApps = [];
  let editingAppIndex = -1;

  function switchProjectModalTab(tabName) {
    const isDeploy = tabName === 'deploy';
    if (tabBtnBasic) tabBtnBasic.classList.toggle('active', !isDeploy);
    if (tabBtnDeploy) tabBtnDeploy.classList.toggle('active', isDeploy);
    if (tabPaneBasic) tabPaneBasic.style.display = isDeploy ? 'none' : 'block';
    if (tabPaneDeploy) tabPaneDeploy.style.display = isDeploy ? 'block' : 'none';
  }

  if (tabBtnBasic) {
    tabBtnBasic.addEventListener('click', function () {
      switchProjectModalTab('basic');
    });
  }
  if (tabBtnDeploy) {
    tabBtnDeploy.addEventListener('click', function () {
      editingAppIndex = -1;
      switchProjectModalTab('deploy');
      renderDeployAppsEditor();
    });
  }

  function genShortId(prefix) {
    return (prefix || 'id') + '-' + Math.random().toString(36).substring(2, 7) + Date.now().toString(36).slice(-3);
  }

  function createDefaultEnv(envName) {
    return {
      id: genShortId('env'),
      env_name: envName || '生产环境',
      method: 'ssh',
      host: '',
      port: 22,
      username: 'root',
      auth_type: 'password',
      password: '',
      ssh_key_path: '',
      ssh_key_content: '',
      remote_dir: '',
      custom_prompt: '',
      ignore_files: '.git\nnode_modules\n.DS_Store'
    };
  }

  function createDefaultDeployApp(appName, appType) {
    return {
      id: genShortId('app'),
      name: appName || '后端API',
      type: appType || 'backend_api',
      envs: [createDefaultEnv('生产环境')]
    };
  }

  function getAppTypeLabel(type) {
    switch (type) {
      case 'web_h5': return '🌐 前端 H5 / Web';
      case 'backend_api': return '⚙️ 后端 API / 服务';
      case 'android_apk': return '📱 APK / App打包';
      case 'docker': return '🐳 Docker 容器';
      default: return '🔧 自定义部署';
    }
  }

  async function testDeployEnvConnection(env, btnEl, statusEl) {
    if (!btnEl) return;
    const origText = btnEl.textContent;
    btnEl.disabled = true;
    btnEl.textContent = '⏳ 测试中...';
    if (statusEl) {
      statusEl.textContent = '正在测试连接...';
      statusEl.style.color = '#8b949e';
    }
    try {
      const res = await fetch('/api/projects/deploy-test', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(env)
      });
      const data = await res.json();
      if (statusEl) {
        statusEl.textContent = data.message || (data.ok ? '✅ 连接成功' : '❌ 连接失败');
        statusEl.style.color = data.ok ? '#3fb950' : '#ff7b72';
      } else {
        alert(data.message || (data.ok ? '✅ 连接成功' : '❌ 连接失败'));
      }
    } catch (err) {
      if (statusEl) {
        statusEl.textContent = '❌ 请求异常: ' + err.message;
        statusEl.style.color = '#ff7b72';
      } else {
        alert('❌ 请求异常: ' + err.message);
      }
    } finally {
      btnEl.disabled = false;
      btnEl.textContent = origText;
    }
  }

  function renderDeployAppsEditor() {
    if (!deployAppsContainer) return;
    deployAppsContainer.innerHTML = '';

    if (tabBtnDeploy) {
      let totalEnvs = 0;
      editingDeployApps.forEach(a => { totalEnvs += (a.envs ? a.envs.length : 0); });
      tabBtnDeploy.textContent = editingDeployApps.length > 0
        ? `🚀 部署配置 (${editingDeployApps.length}个应用 / ${totalEnvs}个环境)`
        : '🚀 部署配置 (多应用/多环境)';
    }

    // 列表模式：展示所有部署配置项，支持连接测试、编辑和删除
    if (editingAppIndex < 0 || editingAppIndex >= editingDeployApps.length) {
      editingAppIndex = -1;
      if (!editingDeployApps || editingDeployApps.length === 0) {
        deployAppsContainer.innerHTML = '<div style="color:#8b949e;font-size:13px;padding:24px 0;text-align:center;">暂无部署配置，请点击右上角「+ 添加部署应用」</div>';
        return;
      }

      editingDeployApps.forEach((app, appIdx) => {
        if (!Array.isArray(app.envs)) app.envs = [];
        const row = document.createElement('div');
        row.className = 'deploy-list-row';

        row.innerHTML = `
          <div class="deploy-list-info">
            <div class="deploy-list-title">
              <span>📦 ${escapeHtml(app.name || '未命名应用')}</span>
              <span class="session-badge status-blue">${escapeHtml(getAppTypeLabel(app.type))}</span>
            </div>
            <div class="deploy-list-envs"></div>
          </div>
          <div class="deploy-list-actions">
            <button type="button" class="btn btn-sm btn-secondary btn-edit-deploy-app">✏️ 编辑</button>
            <button type="button" class="btn btn-sm btn-danger btn-del-deploy-app">🗑️ 删除</button>
          </div>
        `;

        const envsWrap = row.querySelector('.deploy-list-envs');
        if (app.envs.length === 0) {
          envsWrap.innerHTML = '<span style="color:#8b949e;font-size:12px;">暂无环境</span>';
        } else {
          app.envs.forEach(env => {
            const methodTag = (env.method || 'ssh').toUpperCase();
            const hostStr = env.method === 'local_cmd' ? '本地' : (env.host ? `${env.host}:${env.port || 22}` : '未填地址');
            const pill = document.createElement('span');
            pill.className = 'deploy-env-pill';
            pill.innerHTML = `
              <span>${escapeHtml(env.env_name || '环境')} (${escapeHtml(methodTag)}: ${escapeHtml(hostStr)})</span>
              <button type="button" class="btn-inline-test" title="测试该环境连接可用性">🔌 测试</button>
            `;
            const btnTest = pill.querySelector('.btn-inline-test');
            btnTest.addEventListener('click', function () {
              testDeployEnvConnection(env, btnTest, null);
            });
            envsWrap.appendChild(pill);
          });
        }

        row.querySelector('.btn-edit-deploy-app').addEventListener('click', function () {
          editingAppIndex = appIdx;
          renderDeployAppsEditor();
        });

        row.querySelector('.btn-del-deploy-app').addEventListener('click', function () {
          if (!confirm(`确定删除部署配置「${app.name || '未命名'}」吗？`)) return;
          editingDeployApps.splice(appIdx, 1);
          renderDeployAppsEditor();
        });

        deployAppsContainer.appendChild(row);
      });
      return;
    }

    // 单项编辑模式：编辑选中的部署应用及其多环境配置
    const appIdx = editingAppIndex;
    const app = editingDeployApps[appIdx];
    if (!Array.isArray(app.envs)) app.envs = [];

    const card = document.createElement('div');
    card.className = 'deploy-app-card';

    card.innerHTML = `
      <div class="deploy-app-header">
        <div class="deploy-app-title-group">
          <span style="font-size:13px;font-weight:600;color:#58a6ff;">📦 应用名称:</span>
          <input type="text" class="form-input app-name-input" placeholder="如: 前端H5 / 后端API / APK打包" value="${escapeHtml(app.name || '')}">
          <select class="form-select app-type-select">
            <option value="web_h5" ${app.type === 'web_h5' ? 'selected' : ''}>🌐 前端 H5 / Web</option>
            <option value="backend_api" ${(!app.type || app.type === 'backend_api') ? 'selected' : ''}>⚙️ 后端 API / 服务</option>
            <option value="android_apk" ${app.type === 'android_apk' ? 'selected' : ''}>📱 移动端 APK / App打包</option>
            <option value="docker" ${app.type === 'docker' ? 'selected' : ''}>🐳 Docker 容器服务</option>
            <option value="custom" ${app.type === 'custom' ? 'selected' : ''}>🔧 自定义构建部署</option>
          </select>
        </div>
        <div class="deploy-app-actions">
          <button type="button" class="btn btn-sm btn-primary btn-add-env">+ 添加环境</button>
          <button type="button" class="btn btn-sm btn-secondary btn-back-list">✓ 返回列表</button>
        </div>
      </div>
      <div class="deploy-envs-container"></div>
    `;

    card.querySelector('.app-name-input').addEventListener('input', function () {
      app.name = this.value;
    });

    card.querySelector('.app-type-select').addEventListener('change', function () {
      app.type = this.value;
    });

    card.querySelector('.btn-add-env').addEventListener('click', function () {
      const defaultNames = ['生产环境', '测试环境', '本地测试', '预发环境'];
      const nextName = defaultNames[app.envs.length] || `环境-${app.envs.length + 1}`;
      app.envs.push(createDefaultEnv(nextName));
      renderDeployAppsEditor();
    });

    card.querySelector('.btn-back-list').addEventListener('click', function () {
      editingAppIndex = -1;
      renderDeployAppsEditor();
    });

    const envsContainer = card.querySelector('.deploy-envs-container');
    app.envs.forEach((env, envIdx) => {
      const envBox = document.createElement('div');
      envBox.className = 'deploy-env-card';
      const isLocal = env.method === 'local_cmd';
      const isKeyAuth = env.auth_type === 'ssh_key';
      const defaultPort = env.method === 'sftp_ftp' ? 21 : 22;

      envBox.innerHTML = `
        <div class="deploy-env-header">
          <div class="deploy-env-title">
            <span class="deploy-env-badge">环境 ${envIdx + 1}</span>
            <input type="text" class="form-input env-name-input" style="width:140px;padding:4px 8px;font-size:12px;" placeholder="如: 生产环境 / 测试环境" value="${escapeHtml(env.env_name || '')}">
            <select class="form-select env-deploy-type-select" style="width:170px;padding:4px 8px;font-size:12px;">
              <option value="ssh" ${(!env.method || env.method === 'ssh') ? 'selected' : ''}>🖥️ SSH / SCP 远程服务器</option>
              <option value="sftp_ftp" ${env.method === 'sftp_ftp' ? 'selected' : ''}>📁 FTP / SFTP 服务器上传</option>
              <option value="local_cmd" ${env.method === 'local_cmd' ? 'selected' : ''}>💻 仅本地构建 / 脚本部署</option>
            </select>
            <button type="button" class="btn btn-sm btn-secondary btn-test-env" style="padding:3px 9px;font-size:11.5px;">🔌 测试连接</button>
            <span class="env-test-status" style="font-size:11.5px;"></span>
          </div>
          <button type="button" class="btn btn-sm btn-danger btn-del-env" style="padding:2px 8px;font-size:11px;">删除环境</button>
        </div>

        <div class="deploy-env-grid">
          <div class="form-item env-remote-field" style="${isLocal ? 'display:none;' : ''}">
            <label>服务器地址 (Host / IP):</label>
            <input type="text" class="form-input env-host" placeholder="例如: 192.168.1.100" value="${escapeHtml(env.host || '')}">
          </div>
          <div class="form-item env-remote-field" style="${isLocal ? 'display:none;' : ''}">
            <label>端口 (Port):</label>
            <input type="number" class="form-input env-port" placeholder="${defaultPort}" value="${env.port || defaultPort}">
          </div>
          <div class="form-item env-remote-field" style="${isLocal ? 'display:none;' : ''}">
            <label>登录用户名 (Username):</label>
            <input type="text" class="form-input env-username" placeholder="root / ubuntu" value="${escapeHtml(env.username || '')}">
          </div>

          <div class="form-item env-remote-field" style="${isLocal ? 'display:none;' : ''}">
            <label>认证方式:</label>
            <select class="form-select env-auth-type">
              <option value="password" ${(!env.auth_type || env.auth_type === 'password') ? 'selected' : ''}>🔑 密码 / Token</option>
              <option value="ssh_key" ${env.auth_type === 'ssh_key' ? 'selected' : ''}>📜 SSH 证书 / 私钥</option>
            </select>
          </div>
          <div class="form-item env-remote-field env-pwd-wrap" style="${(isLocal || isKeyAuth) ? 'display:none;' : ''}">
            <label>服务器密码 / Token:</label>
            <input type="password" class="form-input env-password" autocomplete="new-password" placeholder="输入密码或Token" value="${escapeHtml(env.password || '')}">
          </div>
          <div class="form-item env-remote-field env-key-wrap" style="${(isLocal || !isKeyAuth) ? 'display:none;' : ''}">
            <label>SSH 私钥文件路径 (可选):</label>
            <input type="text" class="form-input env-ssh-key" placeholder="如 ~/.ssh/id_rsa (或直接在下方粘贴证书文本)" value="${escapeHtml(env.ssh_key_path || '')}">
          </div>
          <div class="form-item env-remote-field" style="${isLocal ? 'display:none;' : ''}">
            <label>远程目标目录 (Remote Dir):</label>
            <input type="text" class="form-input env-remote-dir" placeholder="例如: /var/www/html" value="${escapeHtml(env.remote_dir || '')}">
          </div>

          <div class="form-item span-3 env-remote-field env-key-content-wrap" style="${(isLocal || !isKeyAuth) ? 'display:none;' : ''}">
            <label>SSH 证书/私钥文本 (支持直接粘贴私钥内容 -----BEGIN ... PRIVATE KEY-----):</label>
            <textarea class="form-input env-ssh-key-content" rows="3" placeholder="-----BEGIN OPENSSH PRIVATE KEY-----&#10;可直接将私钥文件内容粘贴在此处，系统将自动以 0600 权限加密隔离保存...">${escapeHtml(env.ssh_key_content || '')}</textarea>
          </div>

          <div class="form-item span-2">
            <label>自定义提示词 (Custom Prompt):</label>
            <textarea class="form-input env-custom-prompt" rows="3" placeholder="输入部署该环境时的专属提示词要求（如：先执行 npm run build 编译，再上传 dist 目录并重启服务）...">${escapeHtml(env.custom_prompt || '')}</textarea>
          </div>
          <div class="form-item">
            <label>忽略文件 (每行一个):</label>
            <textarea class="form-input env-ignore-files" rows="3" placeholder=".git&#10;node_modules&#10;.env">${escapeHtml(env.ignore_files || '')}</textarea>
          </div>
        </div>
      `;

      envBox.querySelector('.env-name-input').addEventListener('input', function () { env.env_name = this.value; });
      envBox.querySelector('.env-deploy-type-select').addEventListener('change', function () {
        env.method = this.value;
        if (this.value === 'sftp_ftp' && (!env.port || env.port === 22)) env.port = 21;
        if (this.value === 'ssh' && (!env.port || env.port === 21)) env.port = 22;
        renderDeployAppsEditor();
      });
      envBox.querySelector('.env-host').addEventListener('input', function () { env.host = this.value.trim(); });
      envBox.querySelector('.env-port').addEventListener('input', function () { env.port = parseInt(this.value, 10) || 22; });
      envBox.querySelector('.env-username').addEventListener('input', function () { env.username = this.value.trim(); });
      envBox.querySelector('.env-auth-type').addEventListener('change', function () {
        env.auth_type = this.value;
        renderDeployAppsEditor();
      });
      envBox.querySelector('.env-password').addEventListener('input', function () { env.password = this.value; });
      envBox.querySelector('.env-ssh-key').addEventListener('input', function () { env.ssh_key_path = this.value.trim(); });
      envBox.querySelector('.env-ssh-key-content').addEventListener('input', function () { env.ssh_key_content = this.value; });
      envBox.querySelector('.env-remote-dir').addEventListener('input', function () { env.remote_dir = this.value.trim(); });
      envBox.querySelector('.env-custom-prompt').addEventListener('input', function () { env.custom_prompt = this.value; });
      envBox.querySelector('.env-ignore-files').addEventListener('input', function () { env.ignore_files = this.value; });

      const btnTestEnv = envBox.querySelector('.btn-test-env');
      const statusSpan = envBox.querySelector('.env-test-status');
      btnTestEnv.addEventListener('click', function () {
        testDeployEnvConnection(env, btnTestEnv, statusSpan);
      });

      envBox.querySelector('.btn-del-env').addEventListener('click', function () {
        app.envs.splice(envIdx, 1);
        renderDeployAppsEditor();
      });

      envsContainer.appendChild(envBox);
    });

    deployAppsContainer.appendChild(card);
  }

  if (btnAddDeployApp) {
    btnAddDeployApp.addEventListener('click', function () {
      const presetNames = ['前端H5', '后端API', 'APK打包'];
      const presetTypes = ['web_h5', 'backend_api', 'android_apk'];
      const idx = editingDeployApps.length;
      const defaultName = presetNames[idx] || `应用-${idx + 1}`;
      const defaultType = presetTypes[idx] || 'backend_api';
      editingDeployApps.push(createDefaultDeployApp(defaultName, defaultType));
      editingAppIndex = editingDeployApps.length - 1;
      switchProjectModalTab('deploy');
      renderDeployAppsEditor();
    });
  }

  function normalizeDirClient(dir) {
    if (!dir) return '';
    return dir.trim().replace(/[\\/]+$/, '');
  }

  function renderProjectsList(projectsStatus, activeProjectId) {
    if (!projectsListContainer) return;
    projectsListContainer.innerHTML = '';
    const dict = I18N[currentLang] || I18N.en;
    const totalCount = projectsStatus ? projectsStatus.length : 0;

    (projectsStatus || []).forEach(p => {
      const isActive = p.is_active || (activeProjectId && p.id === activeProjectId);
      const isRunning = !!p.is_running;
      const row = document.createElement('div');
      row.className = 'project-item-row' + (isActive ? ' active-project' : '');

      const runBadgeClass = isRunning ? 'status-green' : 'status-gray';
      const runBadgeText = isRunning
        ? `${dict.status_running_badge || 'Running'} (PID: ${p.pid || '--'} | Clients: ${p.clients_count || 0})`
        : (dict.badge_standby || 'Standby');
      const activeBadgeHtml = isActive
        ? `<span class="session-badge status-active-star">${dict.badge_active_recent || '⭐ Recent Active'}</span>`
        : '';
      const approveTag = p.auto_approve ? ' [⚡Auto]' : '';
      const deployAppsCount = (p.deploy_apps && p.deploy_apps.length) || p.deploy_apps_count || 0;
      const deployBadgeHtml = deployAppsCount > 0
        ? `<span class="session-badge status-green" title="已配置 ${deployAppsCount} 个部署应用 (密码安全隔离在 ~/.tim-agent/vault)">🚀 部署应用: ${deployAppsCount}</span>`
        : '';

      row.innerHTML = `
        <div class="project-item-main">
          <div class="project-item-title-line">
            <span class="project-item-name">📂 ${escapeHtml(p.name || p.id)}</span>
            ${activeBadgeHtml}
            <span class="session-badge ${runBadgeClass}">${escapeHtml(runBadgeText)}</span>
            <span class="session-badge status-blue">🤖 Agent: ${escapeHtml((p.active_agent || p.agent || 'claude') + approveTag)}</span>
            ${deployBadgeHtml}
          </div>
          <div class="project-item-dir">📁 WorkDir: ${escapeHtml(p.work_dir || '')}</div>
        </div>
        <div class="project-item-actions">
          <button type="button" class="btn btn-sm btn-primary btn-proj-open">${dict.btn_switch_open || '🚀 Open Terminal'}</button>
          <button type="button" class="btn btn-sm btn-secondary btn-proj-activate">${dict.btn_set_active || '▶️ Activate'}</button>
          <button type="button" class="btn btn-sm btn-secondary btn-proj-restart">${dict.btn_restart_proj || '🔄 Restart'}</button>
          <button type="button" class="btn btn-sm btn-secondary btn-proj-edit">${dict.btn_edit_proj || '✏️ Edit'}</button>
          ${totalCount > 1 ? `<button type="button" class="btn btn-sm btn-danger btn-proj-del">${dict.btn_del_proj || '🗑️ Delete'}</button>` : ''}
        </div>
      `;

      // 1. 切换并打开该项目终端 (之前会话保持后台运行)
      const btnOpen = row.querySelector('.btn-proj-open');
      if (btnOpen) {
        btnOpen.onclick = async function () {
          try {
            await fetch('/api/projects/select', {
              method: 'POST',
              headers: { 'Content-Type': 'application/json' },
              body: JSON.stringify({ project_id: p.id })
            });
          } catch (_) {}
          window.open('/terminal?session=' + encodeURIComponent(p.id), '_blank');
          fetchStatus();
        };
      }

      // 2. 启动/设为当前活动项目
      const btnActivate = row.querySelector('.btn-proj-activate');
      if (btnActivate) {
        btnActivate.onclick = async function () {
          try {
            const res = await fetch('/api/projects/select', {
              method: 'POST',
              headers: { 'Content-Type': 'application/json' },
              body: JSON.stringify({ project_id: p.id })
            });
            const resData = await res.json();
            if (resData.ok) {
              fetchStatus();
            } else {
              alert(resData.message || 'Error');
            }
          } catch (_) {
            alert(dict.net_error || 'Network error');
          }
        };
      }

      // 3. 重启该项目会话
      const btnRestart = row.querySelector('.btn-proj-restart');
      if (btnRestart) {
        btnRestart.onclick = async function () {
          const approveText = p.auto_approve ? (dict.auto_approve_yes || '【Auto-Approve】') : (dict.auto_approve_no || '【Standard】');
          if (!confirm(`确定重启项目「${p.name}」的终端会话 (${approveText} [${p.agent}]) 吗？（其他后台项目会话不受影响）`)) return;
          try {
            const res = await fetch('/api/sessions/restart', {
              method: 'POST',
              headers: { 'Content-Type': 'application/json' },
              body: JSON.stringify({
                project_id: p.id,
                agent: p.agent,
                work_dir: p.work_dir,
                auto_approve: p.auto_approve
              })
            });
            const resData = await res.json();
            if (resData.ok) {
              fetchStatus();
            } else {
              alert(resData.message || 'Error');
            }
          } catch (_) {
            alert(dict.net_error || 'Network error');
          }
        };
      }

      // 4. 编辑项目配置
      const btnEdit = row.querySelector('.btn-proj-edit');
      if (btnEdit) {
        btnEdit.onclick = function () {
          openProjectModal(p, 'basic');
        };
      }

      // 5. 删除项目配置
      const btnDel = row.querySelector('.btn-proj-del');
      if (btnDel) {
        btnDel.onclick = async function () {
          if (!confirm(`${dict.confirm_del_proj || 'Delete project '}「${p.name}」?`)) return;
          try {
            const res = await fetch('/api/projects/delete', {
              method: 'POST',
              headers: { 'Content-Type': 'application/json' },
              body: JSON.stringify({ id: p.id })
            });
            const resData = await res.json();
            if (resData.ok) {
              fetchStatus();
            } else {
              alert(resData.message || 'Error');
            }
          } catch (_) {
            alert(dict.net_error || 'Network error');
          }
        };
      }

      projectsListContainer.appendChild(row);
    });
  }

  function escapeHtml(str) {
    return String(str || '')
      .replace(/&/g, '&amp;')
      .replace(/</g, '&lt;')
      .replace(/>/g, '&gt;')
      .replace(/"/g, '&quot;');
  }

  function openProjectModal(proj, defaultTab) {
    if (!projectModal) return;
    const dict = I18N[currentLang] || I18N.en;
    if (projErrorMsg) {
      projErrorMsg.style.display = 'none';
      projErrorMsg.textContent = '';
    }
    // 查找完整配置中的 deploy_apps
    let rawDeployApps = [];
    if (proj && proj.id) {
      if (Array.isArray(proj.deploy_apps)) {
        rawDeployApps = proj.deploy_apps;
      } else if (currentStatus && currentStatus.config && Array.isArray(currentStatus.config.projects)) {
        const foundCfgProj = currentStatus.config.projects.find(item => item.id === proj.id);
        if (foundCfgProj && Array.isArray(foundCfgProj.deploy_apps)) {
          rawDeployApps = foundCfgProj.deploy_apps;
        }
      }
    }
    editingDeployApps = JSON.parse(JSON.stringify(rawDeployApps || []));
    editingAppIndex = -1;
    renderDeployAppsEditor();

    if (proj && proj.id) {
      projectModalTitle.textContent = dict.modal_proj_edit_title || '✏️ 编辑项目配置';
      projIdInput.value = proj.id;
      projNameInput.value = proj.name || '';
      projWorkDirInput.value = proj.work_dir || '';
      if (projAgentSelect) projAgentSelect.value = proj.agent || proj.active_agent || 'claude';
      projAutoApproveCheck.checked = !!proj.auto_approve;
      projSetActiveCheck.checked = !!proj.is_active;
    } else {
      projectModalTitle.textContent = dict.modal_proj_add_title || '📂 新增项目配置';
      projIdInput.value = '';
      projNameInput.value = '';
      projWorkDirInput.value = '';
      if (projAgentSelect) projAgentSelect.value = 'claude';
      projAutoApproveCheck.checked = false;
      projSetActiveCheck.checked = true;
    }
    switchProjectModalTab(defaultTab || 'basic');
    projectModal.style.display = 'flex';
    if (!defaultTab || defaultTab === 'basic') {
      projNameInput.focus();
    }
  }

  function closeProjectModal() {
    if (projectModal) projectModal.style.display = 'none';
  }

  if (btnAddProject) {
    btnAddProject.addEventListener('click', function () {
      openProjectModal(null, 'basic');
    });
  }
  if (btnCloseProjectModal) {
    btnCloseProjectModal.addEventListener('click', closeProjectModal);
  }
  if (btnCancelProjectModal) {
    btnCancelProjectModal.addEventListener('click', closeProjectModal);
  }
  if (projectModal) {
    projectModal.addEventListener('click', function (e) {
      if (e.target === projectModal) closeProjectModal();
    });
  }

  if (projectForm) {
    projectForm.addEventListener('submit', async function (e) {
      e.preventDefault();
      const dict = I18N[currentLang] || I18N.en;
      const id = projIdInput.value.trim();
      const name = projNameInput.value.trim();
      const workDir = projWorkDirInput.value.trim();
      const agent = projAgentSelect ? projAgentSelect.value : 'claude';
      const autoApprove = projAutoApproveCheck.checked;
      const setActive = projSetActiveCheck.checked;

      if (!name || !workDir) {
        switchProjectModalTab('basic');
        if (projErrorMsg) {
          projErrorMsg.textContent = '⚠️ 请填写项目名称与项目工作目录 (WorkDir)';
          projErrorMsg.style.display = 'block';
        }
        return;
      }

      // 前端前置校验：不能同一个目录属于不同项目
      const normInputDir = normalizeDirClient(workDir);
      const existingProjects = (currentStatus && currentStatus.config && currentStatus.config.projects) || [];
      for (const item of existingProjects) {
        if (item.id !== id && normalizeDirClient(item.work_dir) === normInputDir) {
          switchProjectModalTab('basic');
          if (projErrorMsg) {
            projErrorMsg.textContent = `⚠️ ${dict.dup_workdir_err || '工作目录已属于其他项目：'}「${item.name}」(${item.work_dir})`;
            projErrorMsg.style.display = 'block';
          }
          return;
        }
      }

      try {
        const res = await fetch('/api/projects/save', {
          method: 'POST',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify({
            id: id,
            name: name,
            work_dir: workDir,
            agent: agent,
            auto_approve: autoApprove,
            deploy_apps: editingDeployApps,
            set_active: setActive
          })
        });
        const data = await res.json();
        if (data.ok) {
          closeProjectModal();
          fetchStatus();
        } else if (projErrorMsg) {
          projErrorMsg.textContent = '⚠️ ' + (data.message || '保存失败');
          projErrorMsg.style.display = 'block';
        }
      } catch (err) {
        if (projErrorMsg) {
          projErrorMsg.textContent = dict.net_error || 'Network error';
          projErrorMsg.style.display = 'block';
        }
      }
    });
  }

  const btnRegenPwd = document.getElementById('btnRegenPwd');
  btnRegenPwd.addEventListener('click', async function () {
    const confirmMsg = I18N[currentLang].confirm_regen || 'Generate a new random access password?';
    if (!confirm(confirmMsg)) return;
    try {
      const res = await fetch('/api/password/regenerate', { method: 'POST' });
      const data = await res.json();
      if (data.ok) {
        cfgPassword.value = data.password;
        alert((I18N[currentLang].regen_success || 'New password generated: ') + data.password);
        fetchStatus();
      }
    } catch (e) {
      alert(I18N[currentLang].net_error || 'Network error');
    }
  });

  // 重启 Shell 并切换 Agent
  btnRestartPty.addEventListener('click', async function () {
    const selected = cfgAgent.value;
    const workDir = cfgWorkDir ? cfgWorkDir.value.trim() : ((currentStatus && currentStatus.config && currentStatus.config.work_dir) || '');
    const autoApprove = cfgAutoApprove.checked;
    const approveText = autoApprove ? (I18N[currentLang].auto_approve_yes || '【Auto-Approve】') : (I18N[currentLang].auto_approve_no || '【Standard Mode】');
    const confirmPrompt = `${I18N[currentLang].confirm_restart_prefix || 'Restart with '}${approveText} [${selected}]?`;
    if (!confirm(confirmPrompt)) return;
    try {
      const res = await fetch('/api/sessions/restart', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ agent: selected, work_dir: workDir, auto_approve: autoApprove })
      });
      const data = await res.json();
      if (data.ok) {
        alert(I18N[currentLang].restart_success || 'Terminal session restarted and Agent launched!');
        fetchStatus();
      } else {
        alert((I18N[currentLang].restart_fail || 'Failed: ') + (data.message || 'Error'));
      }
    } catch (e) {
      alert(I18N[currentLang].net_error || 'Network error');
    }
  });

  btnRefreshSession.addEventListener('click', fetchStatus);

  // 保存配置
  configForm.addEventListener('submit', async function (e) {
    e.preventDefault();
    saveStatus.textContent = I18N[currentLang].saving_text || 'Saving...';

    const payload = {
      port: parseInt(cfgPort.value, 10),
      password: cfgPassword.value,
      shell: cfgShell.value,
      work_dir: cfgWorkDir ? cfgWorkDir.value.trim() : ((currentStatus && currentStatus.config && currentStatus.config.work_dir) || ''),
      selected_agent: cfgAgent.value,
      auto_approve: cfgAutoApprove.checked,
      theme: cfgTheme ? cfgTheme.value : 'github-dark',
      language: cfgLanguage ? cfgLanguage.value : currentLang,
      keep_alive: cfgKeepAlive.checked,
      remote_access: cfgRemoteAccess ? cfgRemoteAccess.checked : false
    };

    try {
      const res = await fetch('/api/config', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(payload)
      });
      const data = await res.json();
      if (data.ok) {
        if (data.restarted) {
          saveStatus.textContent = (I18N[currentLang].saved_restarted_prefix || 'Settings saved & launched Agent [') + (data.agent || payload.selected_agent) + ']!';
        } else {
          saveStatus.textContent = I18N[currentLang].saved_text || 'Settings saved!';
        }
        setTimeout(() => { saveStatus.textContent = ''; }, 3000);
        fetchStatus();
      } else {
        saveStatus.textContent = (I18N[currentLang].restart_fail || 'Error: ') + data.message;
      }
    } catch (err) {
      saveStatus.textContent = I18N[currentLang].net_error || 'Network error';
    }
  });

  // 版本检测与渲染
  const appVersionBadge = document.getElementById('appVersionBadge');
  const footerVersionText = document.getElementById('footerVersionText');
  const btnCheckUpdate = document.getElementById('btnCheckUpdate');
  const updateStatusText = document.getElementById('updateStatusText');

  let updateChecked = false;

  async function checkForUpdates(manual = false) {
    if (!updateStatusText) return;
    updateStatusText.className = 'update-tip';
    updateStatusText.textContent = I18N[currentLang].checking_update || 'Checking for updates...';

    try {
      const res = await fetch('/api/version/check');
      const data = await res.json();
      if (!data.ok) {
        if (manual) {
          alert((I18N[currentLang].check_failed || 'Update check failed: ') + (data.error || ''));
        }
        updateStatusText.textContent = I18N[currentLang].check_failed || 'Update check failed';
        return;
      }

      if (data.has_update) {
        updateStatusText.className = 'update-tip has-new';
        const newVerNotice = (I18N[currentLang].update_available || 'New version available: ') + data.latest_version;
        updateStatusText.innerHTML = `⭐ <a href="${data.html_url || 'https://github.com/tim-today/tim-agent/releases'}" target="_blank" style="color: #f0883e; text-decoration: underline;">${newVerNotice}</a>`;
        if (btnCheckUpdate) {
          btnCheckUpdate.textContent = `🚀 Update (${data.latest_version})`;
          btnCheckUpdate.classList.remove('btn-secondary');
          btnCheckUpdate.classList.add('btn-primary');
        }
        if (manual) {
          if (confirm(`🎉 ${newVerNotice}\n\n${I18N[currentLang].click_to_update || 'Click to download update'}`)) {
            window.open(data.html_url || 'https://github.com/tim-today/tim-agent/releases', '_blank');
          }
        }
      } else {
        updateStatusText.className = 'update-tip is-latest';
        updateStatusText.textContent = I18N[currentLang].up_to_date || '✓ Up to date';
        if (btnCheckUpdate) {
          btnCheckUpdate.textContent = I18N[currentLang].btn_check_update || '🚀 Update';
        }
        if (manual) {
          alert(I18N[currentLang].up_to_date || 'Currently up to date!');
        }
      }
    } catch (e) {
      updateStatusText.textContent = I18N[currentLang].check_failed || 'Update check failed';
    }
  }

  if (btnCheckUpdate) {
    btnCheckUpdate.addEventListener('click', () => checkForUpdates(true));
  }

  const btnInstallDesktopApp = document.getElementById('btnInstallDesktopApp');
  if (btnInstallDesktopApp) {
    btnInstallDesktopApp.addEventListener('click', async function () {
      btnInstallDesktopApp.disabled = true;
      const origText = btnInstallDesktopApp.textContent;
      btnInstallDesktopApp.textContent = '⏳ 生成中...';
      try {
        const res = await fetch('/api/app/install', { method: 'POST' });
        const data = await res.json();
        if (data.ok) {
          alert(`✅ 已成功生成本机桌面 GUI 应用程序！\n位置: ${data.path}\n\n现在您可以直接用鼠标双击桌面或应用程序列表中的「TimAgent」图标启动，无需打开命令行终端。`);
        } else {
          alert('生成桌面应用失败: ' + (data.message || '未知错误'));
        }
      } catch (e) {
        alert('请求失败: ' + e.message);
      } finally {
        btnInstallDesktopApp.disabled = false;
        btnInstallDesktopApp.textContent = origText;
      }
    });
  }

  // 初始加载
  applyLanguage(currentLang);
  fetchStatus().then(() => {
    if (!updateChecked) {
      updateChecked = true;
      setTimeout(() => checkForUpdates(false), 800);
    }
  });
  setInterval(fetchStatus, 5000);
})();
