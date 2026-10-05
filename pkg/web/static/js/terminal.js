// Tim-Agent 手机定制终端交互逻辑
(function () {
  'use strict';

  // 解析 URL 参数
  const urlParams = new URLSearchParams(window.location.search);
  const sessionId = urlParams.get('session') || 'default';
  const token = urlParams.get('token') || '';

  // 状态与 DOM 元素
  const statusDot = document.getElementById('statusDot');
  const sessionTitle = document.getElementById('sessionTitle');
  const terminalEl = document.getElementById('terminal');
  const mobileKeyboard = document.getElementById('mobileKeyboard');
  const btnCtrlPicker = document.getElementById('btnCtrlPicker');
  const ctrlModal = document.getElementById('ctrlModal');
  const btnCloseCtrlModal = document.getElementById('btnCloseCtrlModal');
  const ctrlAlphaGrid = document.getElementById('ctrlAlphaGrid');
  const btnNumPicker = document.getElementById('btnNumPicker');
  const numModal = document.getElementById('numModal');
  const btnCloseNumModal = document.getElementById('btnCloseNumModal');
  const keyAlt = document.getElementById('keyAlt');
  const commandRow = document.getElementById('commandRow');

  const TERM_I18N = {
    en: {
      term_page_title: "Tim-Agent Terminal",
      term_session_title: "Terminal",
      btn_switch_agent_title: "Switch Agent",
      btn_redraw_title: "Redraw Screen",
      btn_theme_title: "Themes",
      btn_input_title: "Prompt / Text Input Helper",
      btn_zoom_out_title: "Zoom Out",
      btn_zoom_in_title: "Zoom In",
      btn_keyboard_title: "Toggle Keyboard",
      btn_dashboard_title: "Control Center",
      btn_ctrl_picker_title: "Click to select Ctrl+ combination",
      key_backspace_title: "Backspace",
      key_enter_title: "Enter",
      btn_add_cmd: "+ Custom",
      modal_prompt_title: "Prompt / Input Helper",
      btn_paste_clipboard: "📋 Paste Clipboard",
      btn_clear_helper: "Clear",
      btn_send_enter: "Send & Execute (Enter)",
      btn_send_only: "Send Only",
      modal_add_cmd_title: "Add Custom Command",
      label_cmd_name: "Button Name:",
      label_cmd_text: "Command:",
      label_cmd_enter: "Auto Enter",
      btn_cancel: "Cancel",
      btn_save_cmd: "Save",
      modal_agent_title: "Select CLI Agent Preset",
      modal_agent_tip: "Switching will restart session and launch selected Agent in bound workdir",
      label_auto_approve: "⚡ Auto-Approve / Bypass Sandbox",
      modal_ctrl_title: "⌨️ Ctrl+ Shortcuts",
      modal_ctrl_tip: "Click to send shortcut immediately and close modal:",
      ctrl_c_desc: "Interrupt",
      ctrl_d_desc: "EOF / Exit",
      ctrl_z_desc: "Suspend",
      ctrl_l_desc: "Redraw",
      ctrl_r_desc: "Search",
      ctrl_a_desc: "Line Start",
      ctrl_e_desc: "Line End",
      ctrl_u_desc: "Delete Line",
      ctrl_alpha_title: "Other Letters (Ctrl + A~Z):",
      modal_theme_title: "🎨 Terminal Theme",
      modal_theme_tip: "Select comfortable theme for mobile and dark coding:",
      status_online: "Online",
      status_connecting: "Reconnecting...",
      status_offline: "Offline",
      btn_install_pwa: "📲 Install as App (PWA)",
      pwa_ios_tip: "iOS Install: Tap Safari Share button ⎋ at bottom, then choose 'Add to Home Screen'!",
      pwa_installed_tip: "Open browser menu and select 'Install App' or 'Add to Home Screen'.",
      vibes: {
        '帮我修复上面的报错并自测': { text: '⚡ Fix', prompt: 'Fix the error above and verify', title: 'Fix error' },
        '继续': { text: '▶️ Continue', prompt: 'Continue', title: 'Continue execution' },
        '帮我编译构建当前项目并汇报结果': { text: '🛠️ Build', prompt: 'Build and compile project and report results', title: 'Compile & Build' },
        '帮我部署当前项目并检查运行状态': { text: '🚀 Deploy', prompt: 'Deploy the project and check running status', title: 'Deploy project' },
        '运行测试并汇报结果': { text: '🧪 Test', prompt: 'Run tests and report results', title: 'Run tests' },
        '帮我用简明commit message提交当前修改': { text: '📝 Commit', prompt: 'Commit current changes with concise commit message', title: 'Git Commit' }
      }
    },
    zh: {
      term_page_title: "Tim-Agent 终端",
      term_session_title: "终端",
      btn_switch_agent_title: "点击切换当前Agent代理",
      btn_redraw_title: "重绘屏幕 (修复尺寸变动乱码)",
      btn_theme_title: "切换终端配色方案",
      btn_input_title: "Prompt / 长文本输入助手",
      btn_zoom_out_title: "减小字体",
      btn_zoom_in_title: "增大字体",
      btn_keyboard_title: "切换虚拟键盘显隐",
      btn_dashboard_title: "返回控制面板",
      btn_ctrl_picker_title: "点击展开选择字母输入 Ctrl+ 组合键",
      key_backspace_title: "退格删除",
      key_enter_title: "回车换行/执行",
      btn_add_cmd: "+ 自定义",
      modal_prompt_title: "Prompt / 长文本输入助手",
      btn_paste_clipboard: "📋 从剪贴板粘贴",
      btn_clear_helper: "清空",
      btn_send_enter: "发送并执行 (Enter)",
      btn_send_only: "仅发送",
      modal_add_cmd_title: "添加常用快捷命令",
      label_cmd_name: "按钮名称:",
      label_cmd_text: "执行命令:",
      label_cmd_enter: "点击后自动回车执行",
      btn_cancel: "取消",
      btn_save_cmd: "保存按钮",
      modal_agent_title: "选择命令行 Agent 代理",
      modal_agent_tip: "切换后将自动在绑定目录下重新拉起该 Agent",
      label_auto_approve: "⚡ 自动确认 (关闭沙箱 / 免交互弹窗)",
      modal_ctrl_title: "⌨️ Ctrl+ 快捷组合键",
      modal_ctrl_tip: "点击立即发送对应快捷控制符并自动收起：",
      ctrl_c_desc: "中断/取消",
      ctrl_d_desc: "退出/EOF",
      ctrl_z_desc: "挂起后台",
      ctrl_l_desc: "清屏重绘",
      ctrl_r_desc: "反向搜索",
      ctrl_a_desc: "光标行首",
      ctrl_e_desc: "光标行尾",
      ctrl_u_desc: "删除整行",
      ctrl_alpha_title: "其他字母 (Ctrl + A~Z):",
      modal_theme_title: "🎨 终端配色方案",
      modal_theme_tip: "选择适合手机环境与夜晚编码的舒适配色：",
      status_online: "在线正常",
      status_connecting: "正在重连...",
      status_offline: "已离线",
      btn_install_pwa: "📲 安装为本地应用 (PWA)",
      pwa_ios_tip: "iOS 安装方法：点击 Safari 底部「分享」按钮 ⎋，然后选择「添加到主屏幕」即可安装为独立 App！",
      pwa_installed_tip: "可通过浏览器右上角菜单选择「安装应用」或「添加到主屏幕」将终端保存至桌面",
      vibes: {}
    },
    'zh-TW': {
      term_page_title: "Tim-Agent 終端",
      term_session_title: "終端",
      btn_switch_agent_title: "點擊切換當前Agent代理",
      btn_redraw_title: "重繪螢幕 (修復尺寸變動亂碼)",
      btn_theme_title: "切換終端配色主題",
      btn_input_title: "Prompt / 長文字輸入助手",
      btn_zoom_out_title: "縮小字型",
      btn_zoom_in_title: "放大字型",
      btn_keyboard_title: "切換虛擬鍵盤顯示",
      btn_dashboard_title: "返回控制面板",
      btn_ctrl_picker_title: "點擊選擇字母輸入 Ctrl+ 組合鍵",
      key_backspace_title: "倒退刪除",
      key_enter_title: "換行/執行",
      btn_add_cmd: "+ 自訂",
      modal_prompt_title: "Prompt / 長文字輸入助手",
      btn_paste_clipboard: "📋 從剪貼簿貼上",
      btn_clear_helper: "清空",
      btn_send_enter: "傳送並執行 (Enter)",
      btn_send_only: "僅傳送",
      modal_add_cmd_title: "新增常用快捷命令",
      label_cmd_name: "按鈕名稱:",
      label_cmd_text: "執行命令:",
      label_cmd_enter: "點擊後自動回車執行",
      btn_cancel: "取消",
      btn_save_cmd: "儲存按鈕",
      modal_agent_title: "選擇命令列 Agent 代理",
      modal_agent_tip: "切換後將自動在綁定目錄下重新拉起該 Agent",
      label_auto_approve: "⚡ 自動確認 (關閉沙箱 / 免互動)",
      modal_ctrl_title: "⌨️ Ctrl+ 快捷組合鍵",
      modal_ctrl_tip: "點擊立即傳送對應控制符號並自動收起：",
      ctrl_c_desc: "中斷/取消",
      ctrl_d_desc: "結束/EOF",
      ctrl_z_desc: "掛起背景",
      ctrl_l_desc: "清屏重繪",
      ctrl_r_desc: "反向搜尋",
      ctrl_a_desc: "游標行首",
      ctrl_e_desc: "游標行尾",
      ctrl_u_desc: "刪除整行",
      ctrl_alpha_title: "其他字母 (Ctrl + A~Z):",
      modal_theme_title: "🎨 終端配色主題",
      modal_theme_tip: "選擇舒適的配色主題：",
      status_online: "連線正常",
      status_connecting: "重新連線中...",
      status_offline: "已離線",
      btn_install_pwa: "📲 安裝為本機應用程式 (PWA)",
      pwa_ios_tip: "iOS 安裝方式：點擊 Safari 底部「分享」按鈕 ⎋，然後選擇「加入主畫面」即可安裝為獨立 App！",
      pwa_installed_tip: "可透過瀏覽器選單選擇「安裝應用程式」或「加入主畫面」將終端儲存至桌面",
      vibes: {
        '帮我修复上面的报错并自测': { text: '⚡ 修復', prompt: '幫我修復上面的報錯並自測', title: '一鍵修復' },
        '继续': { text: '▶️ 繼續', prompt: '繼續', title: '繼續執行' },
        '帮我编译构建当前项目并汇报结果': { text: '🛠️ 編譯', prompt: '幫我編譯構建當前專案並匯報結果', title: '編譯專案' },
        '帮我部署当前项目并检查运行状态': { text: '🚀 部署', prompt: '幫我部署當前專案並檢查運行狀態', title: '部署專案' },
        '运行测试并汇报结果': { text: '🧪 測試', prompt: '運行測試並匯報結果', title: '跑測試' },
        '帮我用简明commit message提交当前修改': { text: '📝 提交', prompt: '幫我用簡明commit message提交當前修改', title: '提交代碼' }
      }
    },
    ja: {
      term_page_title: "Tim-Agent 端末",
      term_session_title: "端末",
      btn_switch_agent_title: "Agent切り替え",
      btn_redraw_title: "画面再描画",
      btn_theme_title: "配色テーマ",
      btn_input_title: "テキスト入力ヘルパー",
      btn_zoom_out_title: "縮小",
      btn_zoom_in_title: "拡大",
      btn_keyboard_title: "キーボード表示切替",
      btn_dashboard_title: "コントロールセンター",
      btn_ctrl_picker_title: "Ctrl+キーを選択",
      key_backspace_title: "削除",
      key_enter_title: "実行/改行",
      btn_add_cmd: "+ カスタム",
      modal_prompt_title: "テキスト入力ヘルパー",
      btn_paste_clipboard: "📋 クリップボード貼付",
      btn_clear_helper: "クリア",
      btn_send_enter: "送信して実行 (Enter)",
      btn_send_only: "送信のみ",
      modal_add_cmd_title: "カスタムコマンド追加",
      label_cmd_name: "ボタン名:",
      label_cmd_text: "コマンド:",
      label_cmd_enter: "自動Enter",
      btn_cancel: "キャンセル",
      btn_save_cmd: "保存",
      modal_agent_title: "CLI Agent 選択",
      modal_agent_tip: "選択したAgentをワークスペースで再起動します",
      label_auto_approve: "⚡ 自動承認 (サンドボックス解除)",
      modal_ctrl_title: "⌨️ Ctrl+ ショートカット",
      modal_ctrl_tip: "タップして制御コードを送信:",
      ctrl_c_desc: "中断",
      ctrl_d_desc: "終了/EOF",
      ctrl_z_desc: "一時停止",
      ctrl_l_desc: "画面クリア",
      ctrl_r_desc: "履歴検索",
      ctrl_a_desc: "行頭移動",
      ctrl_e_desc: "行末移動",
      ctrl_u_desc: "行削除",
      ctrl_alpha_title: "アルファベット (Ctrl + A~Z):",
      modal_theme_title: "🎨 端末テーマ",
      modal_theme_tip: "快適なテーマを選択してください:",
      status_online: "オンライン",
      status_connecting: "再接続中...",
      status_offline: "オフライン",
      btn_install_pwa: "📲 ホーム画面に追加 (PWA)",
      pwa_ios_tip: "iOS インストール: Safari 画面下の共有ボタン ⎋ をタップし、「ホーム画面に追加」を選択してください。",
      pwa_installed_tip: "ブラウザのメニューから「アプリをインストール」または「ホーム画面に追加」を選択してください。",
      vibes: {
        '帮我修复上面的报错并自测': { text: '⚡ 修正', prompt: 'エラーを修正してテストしてください', title: 'エラー修正' },
        '继续': { text: '▶️ 続行', prompt: '続行してください', title: '続行' },
        '帮我编译构建当前项目并汇报结果': { text: '🛠️ ビルド', prompt: 'プロジェクトをビルドして結果を報告してください', title: 'ビルド' },
        '帮我部署当前项目并检查运行状态': { text: '🚀 デプロイ', prompt: 'プロジェクトをデプロイして稼働確認してください', title: 'デプロイ' },
        '运行测试并汇报结果': { text: '🧪 テスト', prompt: 'テストを実行して結果を報告してください', title: 'テスト実行' },
        '帮我用简明commit message提交当前修改': { text: '📝 コミット', prompt: 'コミットメッセージを作成してコミットしてください', title: 'コミット' }
      }
    },
    ko: {
      term_page_title: "Tim-Agent 터미널",
      term_session_title: "터미널",
      btn_switch_agent_title: "Agent 전환",
      btn_redraw_title: "화면 다시 그리기",
      btn_theme_title: "테마 설정",
      btn_input_title: "프롬프트 입력 도우미",
      btn_zoom_out_title: "글꼴 축소",
      btn_zoom_in_title: "글꼴 확대",
      btn_keyboard_title: "키보드 토글",
      btn_dashboard_title: "제어 센터",
      btn_ctrl_picker_title: "Ctrl+ 단축키 선택",
      key_backspace_title: "지우기",
      key_enter_title: "실행/줄바꿈",
      btn_add_cmd: "+ 사용자 정의",
      modal_prompt_title: "프롬프트 입력 도우미",
      btn_paste_clipboard: "📋 클립보드 붙여넣기",
      btn_clear_helper: "지우기",
      btn_send_enter: "전송 및 실행 (Enter)",
      btn_send_only: "전송만",
      modal_add_cmd_title: "사용자 명령 추가",
      label_cmd_name: "버튼 이름:",
      label_cmd_text: "실행 명령:",
      label_cmd_enter: "자동 Enter",
      btn_cancel: "취소",
      btn_save_cmd: "저장",
      modal_agent_title: "CLI Agent 선택",
      modal_agent_tip: "선택한 Agent로 세션을 재시작합니다",
      label_auto_approve: "⚡ 자동 확인 (샌드박스 비활성화)",
      modal_ctrl_title: "⌨️ Ctrl+ 단축키",
      modal_ctrl_tip: "단축키를 눌러 즉시 전송:",
      ctrl_c_desc: "중단",
      ctrl_d_desc: "종료/EOF",
      ctrl_z_desc: "일시정지",
      ctrl_l_desc: "화면 지우기",
      ctrl_r_desc: "검색",
      ctrl_a_desc: "줄 처음",
      ctrl_e_desc: "줄 끝",
      ctrl_u_desc: "줄 전체 삭제",
      ctrl_alpha_title: "기타 문자 (Ctrl + A~Z):",
      modal_theme_title: "🎨 터미널 테마",
      modal_theme_tip: "편안한 테마를 선택하세요:",
      status_online: "온라인",
      status_connecting: "재연결 중...",
      status_offline: "오프라인",
      btn_install_pwa: "📲 앱으로 설치 (PWA)",
      pwa_ios_tip: "iOS 설치: Safari 하단 공유 버튼 ⎋ 을 누른 후 '홈 화면에 추가'를 선택하세요.",
      pwa_installed_tip: "브라우저 메뉴에서 '앱 설치' 또는 '홈 화면에 추가'를 선택하여 설치할 수 있습니다.",
      vibes: {
        '帮我修复上面的报错并自测': { text: '⚡ 수정', prompt: '오류를 수정하고 자체 테스트를 수행하세요', title: '오류 수정' },
        '继续': { text: '▶️ 계속', prompt: '계속 실행하세요', title: '계속' },
        '帮我编译构建当前项目并汇报结果': { text: '🛠️ 빌드', prompt: '프로젝트를 빌드하고 결과를 보고하세요', title: '빌드' },
        '帮我部署当前项目并检查运行状态': { text: '🚀 배포', prompt: '프로젝트를 배포하고 실행 상태를 확인하세요', title: '배포' },
        '运行测试并汇报结果': { text: '🧪 테스트', prompt: '테스트를 실행하고 결과를 보고하세요', title: '테스트' },
        '帮我用简明commit message提交当前修改': { text: '📝 커밋', prompt: '간결한 커밋 메시지로 커밋하세요', title: '커밋' }
      }
    },
    de: {
      term_page_title: "Tim-Agent Terminal",
      term_session_title: "Terminal",
      btn_switch_agent_title: "Agent wechseln",
      btn_redraw_title: "Bildschirm neu zeichnen",
      btn_theme_title: "Farbschemata",
      btn_input_title: "Eingabe-Assistent",
      btn_zoom_out_title: "Schrift verkleinern",
      btn_zoom_in_title: "Schrift vergrößern",
      btn_keyboard_title: "Tastatur ein-/ausblenden",
      btn_dashboard_title: "Kontrollzentrum",
      btn_ctrl_picker_title: "Ctrl+ Tastenkombination wählen",
      key_backspace_title: "Rücktaste",
      key_enter_title: "Eingabe",
      btn_add_cmd: "+ Benutzerdefiniert",
      modal_prompt_title: "Prompt / Eingabe-Assistent",
      btn_paste_clipboard: "📋 Einfügen",
      btn_clear_helper: "Leeren",
      btn_send_enter: "Senden & Ausführen (Enter)",
      btn_send_only: "Nur senden",
      modal_add_cmd_title: "Befehl hinzufügen",
      label_cmd_name: "Schaltflächenname:",
      label_cmd_text: "Befehl:",
      label_cmd_enter: "Auto-Enter",
      btn_cancel: "Abbrechen",
      btn_save_cmd: "Speichern",
      modal_agent_title: "CLI Agent auswählen",
      modal_agent_tip: "Startet die Sitzung neu mit gewähltem Agent",
      label_auto_approve: "⚡ Automatische Bestätigung",
      modal_ctrl_title: "⌨️ Ctrl+ Tastenkombinationen",
      modal_ctrl_tip: "Kombination direkt senden:",
      ctrl_c_desc: "Abbrechen",
      ctrl_d_desc: "Beenden/EOF",
      ctrl_z_desc: "Anhalten",
      ctrl_l_desc: "Neu zeichnen",
      ctrl_r_desc: "Suchen",
      ctrl_a_desc: "Zeilenanfang",
      ctrl_e_desc: "Zeilenende",
      ctrl_u_desc: "Zeile löschen",
      ctrl_alpha_title: "Buchstaben (Ctrl + A~Z):",
      modal_theme_title: "🎨 Terminal-Farbschema",
      modal_theme_tip: "Farbschema auswählen:",
      status_online: "Online",
      status_connecting: "Verbinde neu...",
      status_offline: "Offline",
      btn_install_pwa: "📲 Als App installieren (PWA)",
      pwa_ios_tip: "iOS-Installation: Tippen Sie in Safari unten auf Teilen ⎋ und wählen Sie 'Zum Home-Bildschirm'.",
      pwa_installed_tip: "Öffnen Sie das Browsermenü und wählen Sie 'App installieren' oder 'Zum Startbildschirm hinzufügen'.",
      vibes: {
        '帮我修复上面的报错并自测': { text: '⚡ Reparieren', prompt: 'Fehler oben beheben und überprüfen', title: 'Fehler beheben' },
        '继续': { text: '▶️ Weiter', prompt: 'Weiter ausführen', title: 'Weiter' },
        '帮我编译构建当前项目并汇报结果': { text: '🛠️ Build', prompt: 'Projekt erstellen und Ergebnis melden', title: 'Projekt bauen' },
        '帮我部署当前项目并检查运行状态': { text: '🚀 Deploy', prompt: 'Projekt bereitstellen und Status prüfen', title: 'Bereitstellen' },
        '运行测试并汇报结果': { text: '🧪 Test', prompt: 'Tests ausführen und Ergebnis melden', title: 'Tests' },
        '帮我用简明commit message提交当前修改': { text: '📝 Commit', prompt: 'Änderungen mit kurzer Commit-Nachricht committen', title: 'Git Commit' }
      }
    }
  };

  let currentUiLang = localStorage.getItem('tim_lang') || 'en';

  function applyTerminalLanguage(lang) {
    if (!TERM_I18N[lang]) lang = 'en';
    currentUiLang = lang;
    localStorage.setItem('tim_lang', lang);
    const dict = TERM_I18N[lang];

    document.querySelectorAll('[data-i18n]').forEach(function (el) {
      const key = el.getAttribute('data-i18n');
      if (dict[key]) {
        el.textContent = dict[key];
      }
    });

    if (dict.btn_add_cmd) {
      const btnAddCmd = document.getElementById('btnAddCmd');
      if (btnAddCmd) btnAddCmd.textContent = dict.btn_add_cmd;
    }

    // 适配 Vibe 按键
    if (dict.vibes) {
      document.querySelectorAll('[data-vibe]').forEach(function (btn) {
        const origVibe = btn.getAttribute('data-vibe');
        if (dict.vibes[origVibe]) {
          btn.textContent = dict.vibes[origVibe].text;
          btn.title = dict.vibes[origVibe].title;
          btn.setAttribute('data-vibe', dict.vibes[origVibe].prompt);
        }
      });
    }

    sessionTitle.textContent = sessionId === 'default' ? dict.term_session_title : sessionId;
  }

  applyTerminalLanguage(currentUiLang);

  let ctrlActive = false;
  let altActive = false;
  let fontSize = parseInt(localStorage.getItem('term_font_size') || '14', 10);

  // 内置终端配色主题方案
  const THEMES = {
    'github-dark': {
      id: 'github-dark',
      name: 'GitHub Dark',
      desc: '经典深灰，层次清晰',
      sample: ['#0d1117', '#3fb950', '#58a6ff', '#bc8cff'],
      ui: { bg: '#0d1117', header: '#161b22', border: '#30363d', btnBg: '#21262d' },
      term: {
        background: '#0d1117',
        foreground: '#c9d1d9',
        cursor: '#58a6ff',
        selectionBackground: '#3b5070',
        black: '#484f58',
        red: '#ff7b72',
        green: '#3fb950',
        yellow: '#d29922',
        blue: '#58a6ff',
        magenta: '#bc8cff',
        cyan: '#39c5cf',
        white: '#b1bac4',
        brightBlack: '#6e7681',
        brightRed: '#ffa198',
        brightGreen: '#56d364',
        brightYellow: '#e3b341',
        brightBlue: '#79c0ff',
        brightMagenta: '#d2a8ff',
        brightCyan: '#56d4dd',
        brightWhite: '#f0f6fc'
      }
    },
    'oled-black': {
      id: 'oled-black',
      name: 'OLED Pure Black',
      desc: '纯黑省电，高对比度',
      sample: ['#000000', '#50fa7b', '#00ff66', '#bd93f9'],
      ui: { bg: '#000000', header: '#0a0a0a', border: '#222222', btnBg: '#141414' },
      term: {
        background: '#000000',
        foreground: '#f0f6fc',
        cursor: '#00ff66',
        selectionBackground: '#264f78',
        black: '#222222',
        red: '#ff5555',
        green: '#50fa7b',
        yellow: '#f1fa8c',
        blue: '#8be9fd',
        magenta: '#bd93f9',
        cyan: '#8be9fd',
        white: '#f8f8f2',
        brightBlack: '#555555',
        brightRed: '#ff6e6e',
        brightGreen: '#69ff94',
        brightYellow: '#ffffa5',
        brightBlue: '#a4ffff',
        brightMagenta: '#d6acff',
        brightCyan: '#a4ffff',
        brightWhite: '#ffffff'
      }
    },
    'dracula': {
      id: 'dracula',
      name: 'Dracula',
      desc: '暗紫风韵，现代极客',
      sample: ['#282a36', '#50fa7b', '#ff79c6', '#bd93f9'],
      ui: { bg: '#282a36', header: '#1e1f29', border: '#44475a', btnBg: '#343746' },
      term: {
        background: '#282a36',
        foreground: '#f8f8f2',
        cursor: '#ff79c6',
        selectionBackground: '#44475a',
        black: '#21222c',
        red: '#ff5555',
        green: '#50fa7b',
        yellow: '#f1fa8c',
        blue: '#bd93f9',
        magenta: '#ff79c6',
        cyan: '#8be9fd',
        white: '#f8f8f2',
        brightBlack: '#6272a4',
        brightRed: '#ff6e6e',
        brightGreen: '#69ff94',
        brightYellow: '#ffffa5',
        brightBlue: '#d6acff',
        brightMagenta: '#ff92df',
        brightCyan: '#a4ffff',
        brightWhite: '#ffffff'
      }
    },
    'one-dark': {
      id: 'one-dark',
      name: 'One Dark',
      desc: '经典代码色，雅致耐看',
      sample: ['#1e1e24', '#98c379', '#61afef', '#c678dd'],
      ui: { bg: '#1e1e24', header: '#18181c', border: '#2c313a', btnBg: '#282c34' },
      term: {
        background: '#1e1e24',
        foreground: '#abb2bf',
        cursor: '#528bff',
        selectionBackground: '#3e4451',
        black: '#282c34',
        red: '#e06c75',
        green: '#98c379',
        yellow: '#e5c07b',
        blue: '#61afef',
        magenta: '#c678dd',
        cyan: '#56b6c2',
        white: '#abb2bf',
        brightBlack: '#5c6370',
        brightRed: '#e88388',
        brightGreen: '#a8d485',
        brightYellow: '#eecf8b',
        brightBlue: '#7cb7ff',
        brightMagenta: '#d68beb',
        brightCyan: '#6dcdd9',
        brightWhite: '#ffffff'
      }
    },
    'monokai': {
      id: 'monokai',
      name: 'Monokai Pro',
      desc: '鲜艳对撞，高辨识度',
      sample: ['#272822', '#a6e22e', '#66d9ef', '#ae81ff'],
      ui: { bg: '#272822', header: '#1e1f1c', border: '#3e3d32', btnBg: '#34352e' },
      term: {
        background: '#272822',
        foreground: '#f8f8f2',
        cursor: '#f8f8f0',
        selectionBackground: '#49483e',
        black: '#272822',
        red: '#f92672',
        green: '#a6e22e',
        yellow: '#f4bf75',
        blue: '#66d9ef',
        magenta: '#ae81ff',
        cyan: '#a1efe4',
        white: '#f8f8f2',
        brightBlack: '#75715e',
        brightRed: '#ff3b7f',
        brightGreen: '#b8ee3b',
        brightYellow: '#ffd866',
        brightBlue: '#78e8ff',
        brightMagenta: '#bf94ff',
        brightCyan: '#b4f9f0',
        brightWhite: '#ffffff'
      }
    },
    'solarized-dark': {
      id: 'solarized-dark',
      name: 'Solarized Dark',
      desc: '复古青绿，柔和舒适',
      sample: ['#002b36', '#859900', '#268bd2', '#d33682'],
      ui: { bg: '#002b36', header: '#00212b', border: '#073642', btnBg: '#073642' },
      term: {
        background: '#002b36',
        foreground: '#839496',
        cursor: '#93a1a1',
        selectionBackground: '#073642',
        black: '#073642',
        red: '#dc322f',
        green: '#859900',
        yellow: '#b58900',
        blue: '#268bd2',
        magenta: '#d33682',
        cyan: '#2aa198',
        white: '#eee8d5',
        brightBlack: '#586e75',
        brightRed: '#cb4b16',
        brightGreen: '#93a1a1',
        brightYellow: '#657b83',
        brightBlue: '#839496',
        brightMagenta: '#6c71c4',
        brightCyan: '#93a1a1',
        brightWhite: '#fdf6e3'
      }
    }
  };

  let currentThemeId = localStorage.getItem('term_theme') || 'github-dark';
  if (!THEMES[currentThemeId]) currentThemeId = 'github-dark';

  function applyTheme(themeId) {
    const themeObj = THEMES[themeId] || THEMES['github-dark'];
    currentThemeId = themeObj.id;
    localStorage.setItem('term_theme', currentThemeId);

    // 应用到 xterm
    if (term) {
      term.options.theme = themeObj.term;
    }

    // 应用到页面 CSS 变量以整体协调
    if (themeObj.ui) {
      document.documentElement.style.setProperty('--bg-color', themeObj.ui.bg);
      document.documentElement.style.setProperty('--header-bg', themeObj.ui.header);
      document.documentElement.style.setProperty('--border-color', themeObj.ui.border);
      document.documentElement.style.setProperty('--btn-bg', themeObj.ui.btnBg);
    }
  }

  // 初始化 xterm
  const initialThemeObj = THEMES[currentThemeId] || THEMES['github-dark'];
  const term = new Terminal({
    cursorBlink: true,
    fontSize: fontSize,
    fontFamily: '"SF Mono", Monaco, Menlo, Consolas, "Courier New", monospace',
    theme: initialThemeObj.term,
    allowProposedApi: true
  });
  applyTheme(currentThemeId);

  const fitAddon = new FitAddon.FitAddon();
  term.loadAddon(fitAddon);
  term.open(terminalEl);
  // 保持 xterm 默认可输入状态，允许在移动端弹出原生软键盘
  if (term.textarea) {
    term.textarea.removeAttribute('inputmode');
    term.textarea.setAttribute('autocapitalize', 'off');
    term.textarea.setAttribute('autocorrect', 'off');
    term.textarea.setAttribute('autocomplete', 'off');
    term.textarea.setAttribute('spellcheck', 'false');
  }
  fitAddon.fit();

  // 点击终端视口或辅助区域唤起系统输入法
  function triggerSystemKeyboard() {
    if (term.textarea) {
      term.textarea.focus();
    }
    term.focus();
  }

  // 移动端点击终端视口唤起输入法
  if (terminalEl) {
    terminalEl.addEventListener('click', function () {
      triggerSystemKeyboard();
    });
  }

  // 监听 Visual Viewport 动态计算软键盘弹出高度，实现虚拟键盘栏贴合输入法键盘顶部
  function setupVisualViewportSync() {
    if (!window.visualViewport) return;

    function handleViewportChange() {
      const vv = window.visualViewport;
      const windowH = window.innerHeight;
      const viewportH = vv.height;
      const offsetTop = vv.offsetTop || 0;

      // 计算被原生键盘推起遮挡的高度
      const keyboardHeight = Math.max(0, windowH - (viewportH + offsetTop));

      const root = document.documentElement;
      if (keyboardHeight > 40) {
        // 软键盘弹起：虚拟键盘上移紧贴在原生输入法顶端
        root.style.setProperty('--kb-offset', keyboardHeight + 'px');
      } else {
        // 软键盘收起：虚拟键盘下移紧贴屏幕底部
        root.style.setProperty('--kb-offset', '0px');
      }

      // 视口变动时重新计算终端自适应
      scheduleResize(false, 60);
    }

    window.visualViewport.addEventListener('resize', handleViewportChange);
    window.visualViewport.addEventListener('scroll', handleViewportChange);
  }

  setupVisualViewportSync();

  // WebSocket 管理
  let ws = null;
  let reconnectTimer = null;
  let isConnecting = false;

  function updateStatus(state) {
    statusDot.className = 'status-dot ' + state;
    if (state === 'online') {
      statusDot.title = '在线正常';
    } else if (state === 'connecting') {
      statusDot.title = '正在重连...';
    } else {
      statusDot.title = '已离线';
    }
  }

  // UTF-8 流解码器，防止中文字符多字节切包截断乱码
  const utf8Decoder = new TextDecoder('utf-8');

  let lastCols = 0;
  let lastRows = 0;
  let resizeDebounceTimer = null;

  function doResize(forceRedraw) {
    if (!ws || ws.readyState !== WebSocket.OPEN) return;
    if (terminalEl.clientWidth < 50 || terminalEl.clientHeight < 50) return;

    try {
      fitAddon.fit();
    } catch (err) {
      return;
    }

    const cols = term.cols;
    const rows = term.rows;

    if (cols < 10 || rows < 3) return;

    if (cols !== lastCols || rows !== lastRows || forceRedraw) {
      lastCols = cols;
      lastRows = rows;
      ws.send(JSON.stringify({
        type: 'resize',
        cols: cols,
        rows: rows
      }));

      // 如果需要重绘，延时发送 \x0c (Ctrl+L) 触发全屏重绘，消除旧尺寸字符覆盖残留
      if (forceRedraw) {
        setTimeout(function () {
          sendInput('\x0c');
        }, 80);
      }
    }
  }

  function scheduleResize(forceRedraw, delay) {
    if (delay === undefined) delay = 120;
    if (resizeDebounceTimer) clearTimeout(resizeDebounceTimer);
    resizeDebounceTimer = setTimeout(function () {
      doResize(forceRedraw);
    }, delay);
  }

  let heartbeatTimer = null;
  let lastMessageTime = Date.now();

  function startHeartbeat() {
    stopHeartbeat();
    lastMessageTime = Date.now();
    heartbeatTimer = setInterval(function () {
      if (!ws || ws.readyState !== WebSocket.OPEN) return;
      // 超过 10 秒未收到任何数据或心跳回应，判定为断线/僵尸连接并触发自愈
      if (Date.now() - lastMessageTime > 10000) {
        console.warn('[TimAgent] 会话心跳超时，强制恢复连接...');
        recoverSession('heartbeat_timeout');
        return;
      }
      try {
        ws.send(JSON.stringify({ type: 'ping' }));
      } catch (err) {
        recoverSession('ping_error');
      }
    }, 3000);
  }

  function stopHeartbeat() {
    if (heartbeatTimer) {
      clearInterval(heartbeatTimer);
      heartbeatTimer = null;
    }
  }

  function connect() {
    if (isConnecting || (ws && ws.readyState === WebSocket.OPEN)) return;
    isConnecting = true;
    updateStatus('connecting');

    const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
    const wsUrl = `${protocol}//${window.location.host}/ws/terminal?session=${encodeURIComponent(sessionId)}&token=${encodeURIComponent(token)}`;

    try {
      ws = new WebSocket(wsUrl);
      ws.binaryType = 'arraybuffer';
    } catch (e) {
      isConnecting = false;
      updateStatus('offline');
      scheduleReconnect();
      return;
    }

    ws.onopen = function () {
      isConnecting = false;
      lastMessageTime = Date.now();
      updateStatus('online');
      startHeartbeat();
      // 连接就绪后自适应并同步尺寸
      scheduleResize(false, 50);
    };

    ws.onmessage = function (event) {
      lastMessageTime = Date.now();
      if (typeof event.data === 'string') {
        try {
          const msg = JSON.parse(event.data);
          if (msg.type === 'pong') return;
        } catch (e) {
          term.write(event.data);
        }
      } else {
        // 使用 UTF-8 流解码，彻底杜绝数据切包造成的中文多字节乱码
        const text = utf8Decoder.decode(event.data, { stream: true });
        term.write(text);
      }
    };

    ws.onclose = function () {
      isConnecting = false;
      stopHeartbeat();
      updateStatus('offline');
      scheduleReconnect();
    };

    ws.onerror = function () {
      isConnecting = false;
      stopHeartbeat();
      updateStatus('offline');
      scheduleReconnect();
    };
  }

  function scheduleReconnect(delay) {
    if (reconnectTimer) clearTimeout(reconnectTimer);
    reconnectTimer = setTimeout(function () {
      connect();
    }, delay !== undefined ? delay : 1500);
  }

  // 核心：会话恢复事件处理 (Session Recovery)
  // 当用户切后台、关掉浏览器再重新打开原网址、锁屏唤醒时，自动快速检测连接并拉起重放恢复
  function recoverSession(triggerReason) {
    console.log('[TimAgent] 触发会话恢复检测, 来源:', triggerReason);
    if (reconnectTimer) clearTimeout(reconnectTimer);

    // 如果连接已不存在、已关闭或正在关闭
    if (!ws || ws.readyState === WebSocket.CLOSED || ws.readyState === WebSocket.CLOSING) {
      isConnecting = false;
      connect();
      return;
    }

    // 如果连接状态看似 OPEN，发一次 ping 探活
    if (ws.readyState === WebSocket.OPEN) {
      try {
        ws.send(JSON.stringify({ type: 'ping' }));
        // 重新同步尺寸以确保终端渲染正常
        scheduleResize(false, 50);
      } catch (e) {
        // 发生写入异常说明是假死 socket，强制销毁并重新连接
        try { ws.close(); } catch (_) {}
        ws = null;
        isConnecting = false;
        connect();
      }
    } else if (ws.readyState === WebSocket.CONNECTING) {
      // 正在连接中，等待即可
    }
  }

  // 监听浏览器切回前台、从 BFCache 恢复、获得焦点、网络重连
  document.addEventListener('visibilitychange', function () {
    if (document.visibilityState === 'visible') {
      recoverSession('visibilitychange');
    }
  });

  window.addEventListener('pageshow', function (e) {
    recoverSession('pageshow' + (e.persisted ? '_bfcache' : ''));
  });

  window.addEventListener('focus', function () {
    recoverSession('window_focus');
  });

  window.addEventListener('online', function () {
    recoverSession('network_online');
  });

  function sendInput(data) {
    if (!ws || ws.readyState !== WebSocket.OPEN) return;

    if (ctrlActive && data.length === 1) {
      const code = data.charCodeAt(0);
      if (code >= 65 && code <= 90) {
        data = String.fromCharCode(code - 64);
      } else if (code >= 97 && code <= 122) {
        data = String.fromCharCode(code - 96);
      }
      setCtrl(false);
    }

    if (altActive) {
      data = '\x1b' + data;
      setAlt(false);
    }

    ws.send(JSON.stringify({
      type: 'input',
      data: data
    }));
  }

  // 监听来自 xterm 本身的键盘输入
  term.onData(function (data) {
    sendInput(data);
  });

  // 窗口改变或手机旋转时优雅防抖调整
  window.addEventListener('resize', function () {
    scheduleResize(false, 150);
  });

  // 手机横竖屏切换时带重绘
  window.addEventListener('orientationchange', function () {
    scheduleResize(true, 300);
  });

  // 一键修复重绘屏幕
  const btnRedrawScreen = document.getElementById('btnRedrawScreen');
  if (btnRedrawScreen) {
    btnRedrawScreen.addEventListener('click', function () {
      term.clear();
      scheduleResize(true, 50);
      term.focus();
    });
  }

  // Ctrl / Alt 粘滞按键逻辑
  function setAlt(active) {
    altActive = active;
    if (keyAlt) keyAlt.classList.toggle('active', altActive);
  }

  if (keyAlt) {
    keyAlt.addEventListener('click', function () {
      setAlt(!altActive);
    });
  }

  // Ctrl+ 组合键选择弹窗与字母发送
  function closeCtrlModal() {
    if (ctrlModal) ctrlModal.style.display = 'none';
  }

  function sendCtrlKey(char) {
    if (!char) return;
    const code = char.toUpperCase().charCodeAt(0) - 64;
    if (code >= 1 && code <= 26) {
      sendInput(String.fromCharCode(code));
    }
    closeCtrlModal();
    term.focus();
  }

  function initCtrlPicker() {
    if (ctrlAlphaGrid) {
      ctrlAlphaGrid.innerHTML = '';
      for (let i = 65; i <= 90; i++) {
        const char = String.fromCharCode(i);
        const btn = document.createElement('button');
        btn.className = 'ctrl-alpha-btn';
        btn.textContent = char;
        btn.title = '发送 Ctrl+' + char;
        btn.onclick = function (e) {
          e.preventDefault();
          sendCtrlKey(char.toLowerCase());
        };
        ctrlAlphaGrid.appendChild(btn);
      }
    }

    document.querySelectorAll('.ctrl-preset-card[data-ctrl]').forEach(function (card) {
      card.onclick = function (e) {
        e.preventDefault();
        const ch = this.getAttribute('data-ctrl');
        sendCtrlKey(ch);
      };
    });
  }

  initCtrlPicker();

  if (btnCtrlPicker) {
    btnCtrlPicker.addEventListener('click', function (e) {
      e.preventDefault();
      if (ctrlModal) ctrlModal.style.display = 'flex';
    });
  }

  if (btnCloseCtrlModal) {
    btnCloseCtrlModal.addEventListener('click', closeCtrlModal);
  }

  if (ctrlModal) {
    ctrlModal.addEventListener('click', function (e) {
      if (e.target === ctrlModal) {
        closeCtrlModal();
      }
    });
  }

  // 0-9 数字选择弹窗逻辑
  function closeNumModal() {
    if (numModal) numModal.style.display = 'none';
  }

  function sendNumKey(num) {
    if (num === undefined || num === null) return;
    sendInput(String(num));
    closeNumModal();
    term.focus();
  }

  if (btnNumPicker) {
    btnNumPicker.addEventListener('click', function (e) {
      e.preventDefault();
      if (numModal) numModal.style.display = 'flex';
    });
  }

  if (btnCloseNumModal) {
    btnCloseNumModal.addEventListener('click', closeNumModal);
  }

  if (numModal) {
    numModal.addEventListener('click', function (e) {
      if (e.target === numModal) {
        closeNumModal();
      }
    });
  }

  document.querySelectorAll('.num-key-btn[data-num]').forEach(function (btn) {
    btn.addEventListener('click', function (e) {
      e.preventDefault();
      const n = this.getAttribute('data-num');
      sendNumKey(n);
    });
  });

  // 虚拟键盘工具栏防失焦处理：点击虚拟按键时阻止 textarea 失去焦点，防止误关闭原生软键盘
  if (mobileKeyboard) {
    mobileKeyboard.addEventListener('pointerdown', function (e) {
      // 允许正常点击按钮，但阻止默认转移焦点 (blur) 行为
      const targetBtn = e.target.closest('button');
      if (targetBtn) {
        e.preventDefault();
      }
    });
  }

  // 虚拟按键点击映射 (除 Backspace 由独立长按/魔法清空处理器接管外)
  document.querySelectorAll('[data-key]').forEach(function (btn) {
    const key = btn.getAttribute('data-key');
    if (key === 'Backspace') return; // 由独立的长按/连删/魔法清空处理器接管

    btn.addEventListener('click', function (e) {
      e.preventDefault();
      switch (key) {
        case 'Escape': sendInput('\x1b'); break;
        case 'Tab': sendInput('\t'); break;
        case 'Enter': sendInput('\r'); break;
        case 'ArrowUp': sendInput('\x1b[A'); break;
        case 'ArrowDown': sendInput('\x1b[B'); break;
        case 'ArrowRight': sendInput('\x1b[C'); break;
        case 'ArrowLeft': sendInput('\x1b[D'); break;
      }
      term.focus();
    });
  });

  // 退格删除键 (Backspace) 魔法增强：
  // 1. 单击：删除单个字符 (\x7f)
  // 2. 长按持续 (>300ms)：开启连续高速删除，每60ms删除一个字符并逐渐加快
  // 3. 魔法长按不放 (>850ms)：触发一键全清魔法！自动发送 Ctrl+U (\x15) 清空当前行所有输入内容，并伴随按键震动与微光视觉反馈
  const btnKeyBackspace = document.getElementById('btnKeyBackspace') || document.querySelector('[data-key="Backspace"]');
  if (btnKeyBackspace) {
    let bsTimer = null;
    let bsInterval = null;
    let bsMagicTimer = null;
    let isLongPress = false;
    let isMagicCleared = false;

    function startBackspaceHold(e) {
      if (e) e.preventDefault();
      isLongPress = false;
      isMagicCleared = false;

      // 立即触发首次单击删除
      sendInput('\x7f');
      term.focus();

      // 300ms 后开始连续退格删除
      bsTimer = setTimeout(function () {
        isLongPress = true;
        let speed = 70; // 连删初始间隔 70ms
        bsInterval = setInterval(function () {
          if (!isMagicCleared) {
            sendInput('\x7f');
          }
        }, speed);
      }, 300);

      // 850ms: 魔法长按触发！一键清空光标前/当前整行全部字符 (Ctrl+U)
      bsMagicTimer = setTimeout(function () {
        isMagicCleared = true;
        isLongPress = true;
        if (bsInterval) {
          clearInterval(bsInterval);
          bsInterval = null;
        }
        // 发送终端通用行清空指令: Ctrl+U (\x15) 及后备退格补充
        sendInput('\x15');
        // 手机震动反馈 (支持的设备)
        if (navigator.vibrate) {
          try { navigator.vibrate([40, 30, 40]); } catch (e) {}
        }
        // 界面魔法视觉动效反馈
        btnKeyBackspace.classList.add('magic-clearing');
        setTimeout(function () {
          btnKeyBackspace.classList.remove('magic-clearing');
        }, 350);
      }, 850);
    }

    function stopBackspaceHold(e) {
      if (bsTimer) {
        clearTimeout(bsTimer);
        bsTimer = null;
      }
      if (bsInterval) {
        clearInterval(bsInterval);
        bsInterval = null;
      }
      if (bsMagicTimer) {
        clearTimeout(bsMagicTimer);
        bsMagicTimer = null;
      }
      btnKeyBackspace.classList.remove('magic-clearing');
      term.focus();
    }

    // 支持触摸与鼠标全场景事件
    btnKeyBackspace.addEventListener('pointerdown', startBackspaceHold);
    btnKeyBackspace.addEventListener('pointerup', stopBackspaceHold);
    btnKeyBackspace.addEventListener('pointerleave', stopBackspaceHold);
    btnKeyBackspace.addEventListener('pointercancel', stopBackspaceHold);
    // 阻止原生 click 避免 pointerdown 已经触发后重复触发
    btnKeyBackspace.addEventListener('click', function (e) {
      e.preventDefault();
    });
  }

  // 斜杠常用命令按键与常用命令绑定
  const btnSlashCmd = document.getElementById('btnSlashCmd');
  if (btnSlashCmd) {
    btnSlashCmd.addEventListener('click', function (e) {
      e.preventDefault();
      sendInput('/');
      triggerSystemKeyboard();
    });
  }

  // 顶部专用输入法唤起按钮 (⌨️)
  const btnOpenInputMethodTop = document.getElementById('btnOpenInputMethodTop');
  if (btnOpenInputMethodTop) {
    btnOpenInputMethodTop.addEventListener('click', function (e) {
      e.preventDefault();
      triggerSystemKeyboard();
    });
  }

  // 虚拟键盘栏输入法唤起按钮 (若存在)
  const btnOpenInputMethod = document.getElementById('btnOpenInputMethod');
  if (btnOpenInputMethod) {
    btnOpenInputMethod.addEventListener('click', function (e) {
      e.preventDefault();
      triggerSystemKeyboard();
    });
  }

  // 多语言适配 Vibe 按键标签与提示词
  if (currentUiLang !== 'zh') {
    const vibeTranslations = {
      '帮我修复上面的报错并自测': { text: '⚡ Fix', prompt: 'Fix the error above and verify', title: 'Fix error' },
      '继续': { text: '▶️ Continue', prompt: 'Continue', title: 'Continue execution' },
      '帮我编译构建当前项目并汇报结果': { text: '🛠️ Build', prompt: 'Build and compile project and report results', title: 'Compile & Build' },
      '帮我部署当前项目并检查运行状态': { text: '🚀 Deploy', prompt: 'Deploy the project and check running status', title: 'Deploy project' },
      '运行测试并汇报结果': { text: '🧪 Test', prompt: 'Run tests and report results', title: 'Run tests' },
      '帮我用简明commit message提交当前修改': { text: '📝 Commit', prompt: 'Commit current changes with concise commit message', title: 'Git Commit' },
    };
    document.querySelectorAll('[data-vibe]').forEach(function (btn) {
      const origVibe = btn.getAttribute('data-vibe');
      if (vibeTranslations[origVibe]) {
        btn.textContent = vibeTranslations[origVibe].text;
        btn.title = vibeTranslations[origVibe].title;
        btn.setAttribute('data-vibe', vibeTranslations[origVibe].prompt);
      }
    });
  }

  // Vibe Coding 专属快捷操作按键监听
  document.querySelectorAll('[data-vibe]').forEach(function (btn) {
    btn.addEventListener('click', function (e) {
      e.preventDefault();
      const vibe = this.getAttribute('data-vibe');
      if (vibe) {
        sendInput(vibe + '\r');
        term.focus();
      }
    });
  });

  // Agent 切换与专属快捷键加载
  const btnSwitchAgent = document.getElementById('btnSwitchAgent');
  const agentNameText = document.getElementById('agentNameText');
  const agentModal = document.getElementById('agentModal');
  const btnCloseAgentModal = document.getElementById('btnCloseAgentModal');
  const agentListGroup = document.getElementById('agentListGroup');
  const modalAutoApprove = document.getElementById('modalAutoApprove');

  let currentAgents = [];
  let currentActiveAgent = 'claude';
  let currentAutoApprove = false;

  async function fetchAgentInfo() {
    try {
      const res = await fetch('/api/status');
      const data = await res.json();
      if (data && data.config) {
        currentAgents = data.config.agents || [];
        currentActiveAgent = data.session && data.session.active_agent ? data.session.active_agent : data.config.selected_agent;
        currentAutoApprove = !!data.config.auto_approve;
        if (modalAutoApprove) {
          modalAutoApprove.checked = currentAutoApprove;
        }
        if (!localStorage.getItem('term_theme') && data.config.theme && THEMES[data.config.theme]) {
          applyTheme(data.config.theme);
        }
        if (data.config.language && data.config.language !== currentUiLang) {
          applyTerminalLanguage(data.config.language);
        }
        updateAgentUI();
      }
    } catch (e) {
      console.warn('获取Agent信息失败', e);
    }
  }

  function updateAgentUI() {
    const found = currentAgents.find(a => a.id === currentActiveAgent);
    const name = found ? found.name : currentActiveAgent;
    const approveBadge = currentAutoApprove ? (currentUiLang === 'zh' ? ' [⚡免确认]' : ' [⚡Auto]') : '';
    agentNameText.textContent = name + approveBadge;
  }

  // 配色方案弹窗与实时切换
  const btnThemeModal = document.getElementById('btnThemeModal');
  const themeModal = document.getElementById('themeModal');
  const btnCloseThemeModal = document.getElementById('btnCloseThemeModal');
  const themeGrid = document.getElementById('themeGrid');

  function renderThemeList() {
    if (!themeGrid) return;
    themeGrid.innerHTML = '';
    Object.keys(THEMES).forEach(key => {
      const t = THEMES[key];
      const card = document.createElement('div');
      card.className = 'theme-card' + (t.id === currentThemeId ? ' active' : '');

      const dotsHtml = t.sample.map(color => `<span class="theme-color-dot" style="background-color: ${color};"></span>`).join('');
      card.innerHTML = `
        <div class="theme-preview-bar">${dotsHtml}</div>
        <div class="theme-card-name">${t.name}</div>
        <div class="theme-card-desc">${t.desc}</div>
      `;

      card.onclick = function () {
        applyTheme(t.id);
        renderThemeList();
        scheduleResize(false, 50);
        closeThemeModal();
      };
      themeGrid.appendChild(card);
    });
  }

  if (btnThemeModal) {
    btnThemeModal.addEventListener('click', function () {
      renderThemeList();
      themeModal.style.display = 'flex';
    });
  }

  function closeThemeModal() {
    if (themeModal) themeModal.style.display = 'none';
  }

  if (btnCloseThemeModal) {
    btnCloseThemeModal.addEventListener('click', closeThemeModal);
  }

  btnSwitchAgent.addEventListener('click', function () {
    renderAgentList();
    agentModal.style.display = 'flex';
  });

  function closeAgentModal() {
    agentModal.style.display = 'none';
  }

  btnCloseAgentModal.addEventListener('click', closeAgentModal);

  function renderAgentList() {
    agentListGroup.innerHTML = '';
    currentAgents.forEach(a => {
      const btn = document.createElement('div');
      btn.className = 'agent-item-btn' + (a.id === currentActiveAgent ? ' active' : '');
      const flagHint = a.auto_approve_flag ? `<span style="color: #e3b341; font-size: 11px;">(支持自动确认: ${a.auto_approve_flag})</span>` : '';
      btn.innerHTML = `
        <div class="agent-item-title">${a.name} (${a.command || 'shell'}) ${flagHint}</div>
        <div class="agent-item-desc">${a.description}</div>
      `;
      btn.onclick = async function () {
        const autoApprove = modalAutoApprove.checked;
        const confirmText = autoApprove ? '【开启自动确认/关闭沙箱】' : '【普通安全确认】';
        if (!confirm(`确定以 ${confirmText} 模式切换到 [${a.name}] 吗？当前终端会话将重启`)) return;
        try {
          const res = await fetch('/api/sessions/restart', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ agent: a.id, auto_approve: autoApprove })
          });
          const resData = await res.json();
          if (resData.ok) {
            currentActiveAgent = a.id;
            currentAutoApprove = autoApprove;
            updateAgentUI();
            closeAgentModal();
            term.clear();
            term.write('\x1b[2J\x1b[H');
            if (ws && ws.readyState === WebSocket.OPEN) {
              ws.close();
            }
          } else {
            alert('切换失败: ' + (resData.message || '未知错误'));
          }
        } catch (err) {
          alert('网络错误');
        }
      };
      agentListGroup.appendChild(btn);
    });
  }

  fetchAgentInfo();

  // Prompt 输入助手模态框
  const inputModal = document.getElementById('inputModal');
  const btnInputHelper = document.getElementById('btnInputHelper');
  const btnCloseModal = document.getElementById('btnCloseModal');
  const helperText = document.getElementById('helperText');
  const btnPasteClipboard = document.getElementById('btnPasteClipboard');
  const btnClearHelper = document.getElementById('btnClearHelper');
  const btnSendWithEnter = document.getElementById('btnSendWithEnter');
  const btnSendOnly = document.getElementById('btnSendOnly');

  btnInputHelper.addEventListener('click', function () {
    inputModal.style.display = 'flex';
    helperText.focus();
  });

  function closeInputModal() {
    inputModal.style.display = 'none';
  }

  btnCloseModal.addEventListener('click', closeInputModal);

  btnPasteClipboard.addEventListener('click', async function () {
    try {
      if (navigator.clipboard && navigator.clipboard.readText) {
        const text = await navigator.clipboard.readText();
        helperText.value += text;
      } else {
        alert('请在文本框内长按粘贴');
      }
    } catch (err) {
      alert('读取剪贴板受限，请直接在文本框长按粘贴');
    }
  });

  btnClearHelper.addEventListener('click', function () {
    helperText.value = '';
  });

  btnSendWithEnter.addEventListener('click', function () {
    const text = helperText.value;
    if (text) {
      sendInput(text + '\r');
      helperText.value = '';
      closeInputModal();
      term.focus();
    }
  });

  btnSendOnly.addEventListener('click', function () {
    const text = helperText.value;
    if (text) {
      sendInput(text);
      helperText.value = '';
      closeInputModal();
      term.focus();
    }
  });

  // 字体大小调节
  document.getElementById('btnZoomIn').addEventListener('click', function () {
    if (fontSize < 24) {
      fontSize += 2;
      term.options.fontSize = fontSize;
      scheduleResize(true, 50);
      localStorage.setItem('term_font_size', fontSize);
    }
  });

  document.getElementById('btnZoomOut').addEventListener('click', function () {
    if (fontSize > 10) {
      fontSize -= 2;
      term.options.fontSize = fontSize;
      scheduleResize(true, 50);
      localStorage.setItem('term_font_size', fontSize);
    }
  });

  // 切换虚拟键盘显隐
  document.getElementById('btnToggleKeyboard').addEventListener('click', function () {
    mobileKeyboard.classList.toggle('collapsed');
    scheduleResize(true, 250);
  });

  // 全屏模式切换 (保留顶部工具栏，调用浏览器标准全屏扩展可视区域)
  const btnToggleFullscreen = document.getElementById('btnToggleFullscreen');

  function toggleBrowserFullscreen() {
    if (!document.fullscreenElement) {
      if (document.documentElement.requestFullscreen) {
        document.documentElement.requestFullscreen().catch(() => {});
      } else if (document.documentElement.webkitRequestFullscreen) {
        document.documentElement.webkitRequestFullscreen();
      }
    } else {
      if (document.exitFullscreen) {
        document.exitFullscreen().catch(() => {});
      } else if (document.webkitExitFullscreen) {
        document.webkitExitFullscreen();
      }
    }
    scheduleResize(true, 150);
  }

  if (btnToggleFullscreen) {
    btnToggleFullscreen.addEventListener('click', toggleBrowserFullscreen);
  }

  // 清除旧版本可能残留的 tim_immersive 缓存，确保顶栏绝对可见
  try {
    localStorage.removeItem('tim_immersive');
  } catch (e) {}

  // 手机端右上角齿轮：关于软件与版本信息弹窗
  const btnAboutModal = document.getElementById('btnAboutModal');
  const aboutModal = document.getElementById('aboutModal');
  const btnCloseAboutModal = document.getElementById('btnCloseAboutModal');
  const appVersionText = document.getElementById('appVersionText');
  const aboutUptime = document.getElementById('aboutUptime');
  const btnCheckUpdateTerminal = document.getElementById('btnCheckUpdateTerminal');

  function openAboutModal() {
    if (aboutModal) {
      aboutModal.style.display = 'flex';
      const isStandalone = window.matchMedia('(display-mode: standalone)').matches || window.navigator.standalone === true;
      if (btnInstallPwaModal) {
        btnInstallPwaModal.style.display = 'block';
        if (isStandalone) {
          btnInstallPwaModal.textContent = '✓ 已作为本地应用运行';
          btnInstallPwaModal.disabled = true;
          btnInstallPwaModal.style.opacity = '0.7';
        } else {
          btnInstallPwaModal.disabled = false;
          btnInstallPwaModal.style.opacity = '1';
        }
      }
      fetch('/api/status')
        .then(res => res.json())
        .then(data => {
          if (appVersionText && data.version) {
            appVersionText.textContent = data.version;
          }
          if (aboutUptime && data.uptime) {
            aboutUptime.textContent = data.uptime;
          }
        })
        .catch(err => console.error('获取软件信息失败:', err));
    }
  }

  function closeAboutModal() {
    if (aboutModal) {
      aboutModal.style.display = 'none';
    }
  }

  if (btnAboutModal) {
    btnAboutModal.addEventListener('click', openAboutModal);
  }
  if (btnCloseAboutModal) {
    btnCloseAboutModal.addEventListener('click', closeAboutModal);
  }
  if (aboutModal) {
    aboutModal.addEventListener('click', function (e) {
      if (e.target === aboutModal) closeAboutModal();
    });
  }

  if (btnCheckUpdateTerminal) {
    btnCheckUpdateTerminal.addEventListener('click', async function () {
      btnCheckUpdateTerminal.disabled = true;
      btnCheckUpdateTerminal.textContent = '检查中...';
      try {
        const res = await fetch('/api/version/check');
        const data = await res.json();
        if (data.ok) {
          if (data.has_update) {
            btnCheckUpdateTerminal.textContent = `发现新版 ${data.latest_version}`;
            if (confirm(`发现 Tim-Agent 新版本 ${data.latest_version}！\n是否前往 GitHub 查看更新？`)) {
              window.open(data.html_url || 'https://github.com/tim-today/tim-agent/releases', '_blank');
            }
          } else {
            btnCheckUpdateTerminal.textContent = '已是最新版 ✓';
            setTimeout(() => {
              btnCheckUpdateTerminal.textContent = '检查更新';
              btnCheckUpdateTerminal.disabled = false;
            }, 2500);
          }
        } else {
          alert('检测版本失败: ' + (data.error || '网络超时'));
          btnCheckUpdateTerminal.textContent = '检查更新';
          btnCheckUpdateTerminal.disabled = false;
        }
      } catch (err) {
        alert('无法连接到 GitHub 版本检测服务');
        btnCheckUpdateTerminal.textContent = '检查更新';
        btnCheckUpdateTerminal.disabled = false;
      }
    });
  }

  // ========================================================
  // PWA (Progressive Web App) 支持与防陈旧缓存自动更新机制
  // 核心：Network-First 策略，自动版本检测与 Controller 更新热接管
  // ========================================================
  const btnInstallPwaTop = document.getElementById('btnInstallPwaTop');
  const btnInstallPwaModal = document.getElementById('btnInstallPwaModal');
  let deferredInstallPrompt = null;

  // 1. 监听浏览器 beforeinstallprompt 事件 (Chrome, Edge, Android 等)
  window.addEventListener('beforeinstallprompt', (e) => {
    e.preventDefault();
    deferredInstallPrompt = e;
    if (btnInstallPwaTop) btnInstallPwaTop.style.display = 'inline-flex';
    if (btnInstallPwaModal) {
      btnInstallPwaModal.style.display = 'block';
      btnInstallPwaModal.disabled = false;
      btnInstallPwaModal.style.opacity = '1';
    }
  });

  // 2. 监听 appinstalled 事件
  window.addEventListener('appinstalled', () => {
    deferredInstallPrompt = null;
    if (btnInstallPwaTop) btnInstallPwaTop.style.display = 'none';
    if (btnInstallPwaModal) {
      btnInstallPwaModal.textContent = '✓ 已成功安装为本地应用';
      btnInstallPwaModal.disabled = true;
      btnInstallPwaModal.style.opacity = '0.7';
    }
    console.log('[PWA] 应用已成功安装到本地');
  });

  // 3. 点击安装应用处理
  async function triggerInstallPwa() {
    const isStandalone = window.matchMedia('(display-mode: standalone)').matches || window.navigator.standalone === true;
    if (isStandalone) {
      alert('当前已作为本地 PWA 独立应用运行！');
      return;
    }

    if (deferredInstallPrompt) {
      deferredInstallPrompt.prompt();
      const { outcome } = await deferredInstallPrompt.userChoice;
      if (outcome === 'accepted') {
        if (btnInstallPwaTop) btnInstallPwaTop.style.display = 'none';
        if (btnInstallPwaModal) {
          btnInstallPwaModal.textContent = '✓ 已成功安装为本地应用';
          btnInstallPwaModal.disabled = true;
          btnInstallPwaModal.style.opacity = '0.7';
        }
      }
      deferredInstallPrompt = null;
    } else {
      // 针对未触发原生 prompt (如 Android 纯 HTTP 局域网访问或 iOS Safari) 给出精准指引
      const isIOS = /iPad|iPhone|iPod/.test(navigator.userAgent) && !window.MSStream;
      const isAndroid = /Android/i.test(navigator.userAgent);
      const dict = TERM_I18N[currentUiLang] || TERM_I18N.en;

      if (isIOS) {
        alert(dict.pwa_ios_tip || 'iOS 安装方法：点击 Safari 底部「分享」按钮 ⎋，然后选择「添加到主屏幕」即可安装为独立 App！');
      } else if (isAndroid) {
        alert(
          '📱 Android Chrome 安装指引：\n' +
          '1. 点击 Chrome 右上角三个点「⋮」菜单\n' +
          '2. 选择「添加到主屏幕」或「安装应用」\n' +
          '💡 荣耀 MagicOS / 华为系统提醒：请确保在手机【设置 -> 应用管理 -> Chrome】中已开启「创建桌面快捷方式」权限！\n\n' +
          '💡 局域网专属提示：因浏览器安全策略，纯 HTTP 局域网地址默认限制自动安装弹窗。若需开启原生 PWA 弹窗，可在 Chrome 访问 chrome://flags/#unsafely-treat-insecure-origin-as-secure 将当前连接地址设为信任。'
        );
      } else {
        alert(dict.pwa_installed_tip || '可通过浏览器右上角菜单选择「安装应用」或「添加到主屏幕」将终端保存至桌面');
      }
    }
  }

  if (btnInstallPwaTop) {
    btnInstallPwaTop.addEventListener('click', triggerInstallPwa);
  }
  if (btnInstallPwaModal) {
    btnInstallPwaModal.addEventListener('click', triggerInstallPwa);
  }

  // 4. 注册 Service Worker 并设立防陈旧缓存自动更新监听
  if ('serviceWorker' in navigator) {
    window.addEventListener('load', () => {
      navigator.serviceWorker.register('/sw.js', { scope: '/' })
        .then(reg => {
          // 监听新 Service Worker 下载与准备阶段
          reg.onupdatefound = () => {
            const installingWorker = reg.installing;
            if (installingWorker) {
              installingWorker.onstatechange = () => {
                if (installingWorker.state === 'installed' && navigator.serviceWorker.controller) {
                  console.log('[PWA] 检测到服务端新版本，通知 Service Worker 立即接管...');
                  installingWorker.postMessage({ action: 'skipWaiting' });
                }
              };
            }
          };
        })
        .catch(err => {
          console.warn('[PWA] ServiceWorker 注册异常 (非致命):', err);
        });

      // 当新版本 Service Worker 激活并取得控制权时，热刷新当前页面以无感同步最新版本
      let swRefreshing = false;
      navigator.serviceWorker.addEventListener('controllerchange', () => {
        if (!swRefreshing) {
          swRefreshing = true;
          console.log('[PWA] 新版本 Controller 接管成功，正在刷新以保持最新版本一致...');
          window.location.reload();
        }
      });
    });

    // 页面切回前台时，主动触发一次 update() 检查服务端 Service Worker 字节变更
    document.addEventListener('visibilitychange', () => {
      if (document.visibilityState === 'visible') {
        navigator.serviceWorker.getRegistration().then(reg => {
          if (reg) {
            reg.update().catch(() => {});
          }
        });
      }
    });
  }

  // 开始连接
  connect();
})();
