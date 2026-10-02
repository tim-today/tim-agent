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
      label_workdir: "安全绑定工作目录 (WorkDir):",
      hint_workdir: "安全隔离：终端与 Agent 命令均在此目录内运行",
      label_agent: "选择命令行 Agent 代理:",
      label_auto_approve: "⚡ 自动确认 / 关闭沙箱 (默认不选)",
      hint_agent: "切换后点击“保存配置”，系统将自动重启主会话并在绑定目录下拉起所选 Agent",
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
  const cfgTheme = document.getElementById('cfgTheme');
  const sessAgent = document.getElementById('sessAgent');
  const sessWorkDir = document.getElementById('sessWorkDir');

  function onLanguageChange(newLang) {
    applyLanguage(newLang);
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
          keep_alive: cfg.keep_alive
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
    cfgPort.value = cfg.port;
    cfgPassword.value = cfg.password || '';
    cfgShell.value = cfg.shell || '';
    cfgWorkDir.value = cfg.work_dir || '';
    cfgAutoApprove.checked = !!cfg.auto_approve;
    if (cfgTheme && cfg.theme) {
      cfgTheme.value = cfg.theme;
    }
    const lang = cfg.language || currentLang || 'en';
    applyLanguage(lang);
    cfgKeepAlive.checked = cfg.keep_alive;

    // 填充 Agent 选项
    if (cfg.agents && cfg.agents.length > 0) {
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
  }

  function renderSession(sess) {
    if (!sess) return;
    sessPid.textContent = `PID: ${sess.pid || '--'}`;
    const clientsPrefix = I18N[currentLang].clients_label || 'Clients:';
    sessClients.textContent = `${clientsPrefix} ${sess.clients_count || 0}`;
    if (sess.active_agent) {
      const approveBadge = currentStatus && currentStatus.config && currentStatus.config.auto_approve ? ' [⚡Auto-Approve]' : '';
      sessAgent.textContent = `Agent: ${sess.active_agent}${approveBadge}`;
    }
    if (sess.work_dir) {
      const workDirPrefix = I18N[currentLang].workdir_label || 'WorkDir:';
      sessWorkDir.textContent = `📁 ${workDirPrefix} ${sess.work_dir}`;
    }
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
    const workDir = cfgWorkDir.value.trim();
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
      work_dir: cfgWorkDir.value.trim(),
      selected_agent: cfgAgent.value,
      auto_approve: cfgAutoApprove.checked,
      theme: cfgTheme ? cfgTheme.value : 'github-dark',
      language: cfgLanguage ? cfgLanguage.value : currentLang,
      keep_alive: cfgKeepAlive.checked
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
