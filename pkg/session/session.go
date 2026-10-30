package session

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"path/filepath"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

// WsMessage WebSocket 交互协议
// Type: "input" (终端输入), "resize" (窗口变化), "ping" (心跳), "output" (输出数据)
type WsMessage struct {
	Type string `json:"type"`
	Data string `json:"data,omitempty"`
	Cols uint16 `json:"cols,omitempty"`
	Rows uint16 `json:"rows,omitempty"`
}

// Session 代表一个独立常驻运行的终端会话
type Session struct {
	ID           string    `json:"id"`
	Name         string    `json:"name"`
	Shell        string    `json:"shell"`
	WorkDir      string    `json:"work_dir"`
	ActiveAgent  string    `json:"active_agent"`
	AutoApprove  bool      `json:"auto_approve"`
	CreatedAt    time.Time `json:"created_at"`
	LastActiveAt time.Time `json:"last_active_at"`
	Pid          int       `json:"pid"`
	IsRunning    bool      `json:"is_running"`
	Cols         uint16    `json:"cols"`
	Rows         uint16    `json:"rows"`
	Lang         string    `json:"lang"`

	pty        PTYSession
	buffer     *RingBuffer
	clients    map[*websocket.Conn]bool
	mu         sync.RWMutex
	stopCh     chan struct{}
	onActivity func(sessID string)
}

func newSession(id, name, shell, workDir, activeAgent, initCmd, lang string) (*Session, error) {
	if id == "" {
		id = "default"
	}
	if name == "" {
		name = "主会话"
	}
	if lang == "" {
		lang = "en"
	}

	cols := uint16(100)
	rows := uint16(30)

	ptySess, err := startPTY(shell, workDir, cols, rows)
	if err != nil {
		return nil, fmt.Errorf("启动虚拟终端失败: %w", err)
	}

	now := time.Now()
	s := &Session{
		ID:           id,
		Name:         name,
		Shell:        shell,
		WorkDir:      workDir,
		ActiveAgent:  activeAgent,
		CreatedAt:    now,
		LastActiveAt: now,
		Pid:          ptySess.Pid(),
		IsRunning:    true,
		Cols:         cols,
		Rows:         rows,
		Lang:         lang,
		pty:          ptySess,
		buffer:       NewRingBuffer(1024 * 1024), // 1MB 输出缓存
		clients:      make(map[*websocket.Conn]bool),
		stopCh:       make(chan struct{}),
	}

	// 启动后台读取协程：读取 PTY 输出，写入 RingBuffer，并向所有 WebSocket 广播
	go s.readLoop()

	// 若指定了代理软件命令 (如 claude/codex/opcode/pi/gemini)，在 PTY 就绪后自动拉起
	if initCmd != "" {
		go func() {
			// 等待 Shell 完成 rc 脚本与环境加载
			time.Sleep(600 * time.Millisecond)
			msg := fmt.Sprintf("\r\n\x1b[32m[TimAgent] Starting Agent: %s (%s)...\x1b[0m\r\n", activeAgent, initCmd)
			switch lang {
			case "zh":
				msg = fmt.Sprintf("\r\n\x1b[32m[TimAgent] 正在启动 Agent: %s (%s)...\x1b[0m\r\n", activeAgent, initCmd)
			case "zh-TW":
				msg = fmt.Sprintf("\r\n\x1b[32m[TimAgent] 正在啟動 Agent: %s (%s)...\x1b[0m\r\n", activeAgent, initCmd)
			case "ja":
				msg = fmt.Sprintf("\r\n\x1b[32m[TimAgent] Agentを起動しています: %s (%s)...\x1b[0m\r\n", activeAgent, initCmd)
			case "ko":
				msg = fmt.Sprintf("\r\n\x1b[32m[TimAgent] Agent 시작 중: %s (%s)...\x1b[0m\r\n", activeAgent, initCmd)
			case "de":
				msg = fmt.Sprintf("\r\n\x1b[32m[TimAgent] Starte Agent: %s (%s)...\x1b[0m\r\n", activeAgent, initCmd)
			}
			s.broadcast(websocket.TextMessage, []byte(msg))
			_ = s.WriteString(initCmd + "\n")
		}()
	}

	return s, nil
}

// readLoop 监听 PTY 输出
func (s *Session) readLoop() {
	buf := make([]byte, 4096)
	for {
		select {
		case <-s.stopCh:
			return
		default:
			n, err := s.pty.Read(buf)
			if err != nil {
				if err != io.EOF {
					log.Printf("[Session %s] PTY 读取结束: %v", s.ID, err)
				}
				s.mu.Lock()
				s.IsRunning = false
				s.mu.Unlock()
				exitMsg := "\r\n\x1b[33m[TimAgent] Terminal process exited\x1b[0m\r\n"
				switch s.Lang {
				case "zh":
					exitMsg = "\r\n\x1b[33m[TimAgent] 终端进程已退出\x1b[0m\r\n"
				case "zh-TW":
					exitMsg = "\r\n\x1b[33m[TimAgent] 終端程序已結束\x1b[0m\r\n"
				case "ja":
					exitMsg = "\r\n\x1b[33m[TimAgent] 端末プロセスが終了しました\x1b[0m\r\n"
				case "ko":
					exitMsg = "\r\n\x1b[33m[TimAgent] 터미널 프로세스가 종료되었습니다\x1b[0m\r\n"
				case "de":
					exitMsg = "\r\n\x1b[33m[TimAgent] Terminal-Prozess beendet\x1b[0m\r\n"
				}
				s.broadcast(websocket.TextMessage, []byte(exitMsg))
				return
			}

			if n > 0 {
				data := buf[:n]
				// 写入环形缓冲区备用重连
				_, _ = s.buffer.Write(data)
				// 广播给所有已连接的前端页面
				s.broadcast(websocket.BinaryMessage, data)
			}
		}
	}
}

func (s *Session) broadcast(msgType int, data []byte) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	for conn := range s.clients {
		if err := conn.WriteMessage(msgType, data); err != nil {
			// 写入失败由各连接的读循环处理断开
		}
	}
}

// Touch 更新会话最近活跃时间，并触发活跃回调
func (s *Session) Touch() {
	s.mu.Lock()
	s.LastActiveAt = time.Now()
	cb := s.onActivity
	id := s.ID
	s.mu.Unlock()
	if cb != nil {
		cb(id)
	}
}

// Attach 手机端或PC端连接该会话（支持断线重新连接并恢复现场）
func (s *Session) Attach(conn *websocket.Conn) {
	s.mu.Lock()
	s.clients[conn] = true
	s.LastActiveAt = time.Now()
	cb := s.onActivity
	sessID := s.ID
	infoPayload := map[string]any{
		"type":         "session_info",
		"id":           s.ID,
		"name":         s.Name,
		"work_dir":     s.WorkDir,
		"active_agent": s.ActiveAgent,
		"auto_approve": s.AutoApprove,
	}
	s.mu.Unlock()

	if cb != nil {
		cb(sessID)
	}

	log.Printf("[Session %s (%s)] 客户端已连接 (当前连接数: %d)", s.ID, s.Name, s.ClientsCount())

	// 0. 发送当前会话元信息 (前端据此同步顶部项目名与 Agent 状态)
	if infoBytes, err := json.Marshal(infoPayload); err == nil {
		_ = conn.WriteMessage(websocket.TextMessage, infoBytes)
	}

	// 1. 发送最近的历史输出回放，秒级恢复屏幕现场
	snapshot := s.buffer.Snapshot()
	if len(snapshot) > 0 {
		_ = conn.WriteMessage(websocket.BinaryMessage, snapshot)
	}

	// 2. 如果终端已退出，提醒
	if !s.IsRunning {
		termMsg := "\r\n\x1b[31m[TimAgent] Current session has ended, you can restart it in the control panel\x1b[0m\r\n"
		switch s.Lang {
		case "zh":
			termMsg = "\r\n\x1b[31m[TimAgent] 当前会话已终止，可在控制台或顶部切换中重启\x1b[0m\r\n"
		case "zh-TW":
			termMsg = "\r\n\x1b[31m[TimAgent] 目前會話已終止，可在控制台重新啟動\x1b[0m\r\n"
		case "ja":
			termMsg = "\r\n\x1b[31m[TimAgent] セッションが終了しました。管理パネルから再起動できます\x1b[0m\r\n"
		case "ko":
			termMsg = "\r\n\x1b[31m[TimAgent] 세션이 종료되었습니다. 제어판에서 재시작할 수 있습니다\x1b[0m\r\n"
		case "de":
			termMsg = "\r\n\x1b[31m[TimAgent] Sitzung beendet, kann im Kontrollzentrum neu gestartet werden\x1b[0m\r\n"
		}
		_ = conn.WriteMessage(websocket.TextMessage, []byte(termMsg))
	}

	// 3. 处理客户端发来的输入和调整尺寸命令
	defer func() {
		s.mu.Lock()
		delete(s.clients, conn)
		rem := len(s.clients)
		s.mu.Unlock()
		_ = conn.Close()
		log.Printf("[Session %s (%s)] 客户端已断开 (剩余连接数: %d)", s.ID, s.Name, rem)
	}()

	for {
		messageType, p, err := conn.ReadMessage()
		if err != nil {
			break
		}

		if messageType == websocket.TextMessage {
			var msg WsMessage
			if err := json.Unmarshal(p, &msg); err == nil {
				switch msg.Type {
				case "input":
					s.Touch()
					_, _ = s.pty.Write([]byte(msg.Data))
				case "resize":
					if msg.Cols >= 10 && msg.Rows >= 3 {
						s.mu.Lock()
						if s.Cols != msg.Cols || s.Rows != msg.Rows {
							s.Cols = msg.Cols
							s.Rows = msg.Rows
							_ = s.pty.Resize(msg.Cols, msg.Rows)
						}
						s.mu.Unlock()
					}
				case "ping":
					_ = conn.WriteJSON(WsMessage{Type: "pong"})
				}
			} else {
				// 普通纯文本输入
				s.Touch()
				_, _ = s.pty.Write(p)
			}
		} else if messageType == websocket.BinaryMessage {
			s.Touch()
			_, _ = s.pty.Write(p)
		}
	}
}

func (s *Session) ClientsCount() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.clients)
}

func (s *Session) ListClients() []string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	var list []string
	for c := range s.clients {
		list = append(list, c.RemoteAddr().String())
	}
	return list
}

func (s *Session) WriteString(str string) error {
	_, err := s.pty.Write([]byte(str))
	return err
}

func (s *Session) Close() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.IsRunning {
		return nil
	}

	close(s.stopCh)
	s.IsRunning = false
	return s.pty.Close()
}

// Manager 全局终端会话管理器（支持多项目会话并行常驻后台）
type Manager struct {
	sessions     map[string]*Session
	lastActiveID string
	onActiveFunc func(id string)
	mu           sync.RWMutex
}

var globalManager = &Manager{
	sessions: make(map[string]*Session),
}

func GetManager() *Manager {
	return globalManager
}

// SetOnActivityCallback 注册会话活跃度变更回调（用于同步记录最近操作的项目）
func (m *Manager) SetOnActivityCallback(fn func(id string)) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.onActiveFunc = fn
}

func (m *Manager) recordActivity(id string) {
	m.mu.Lock()
	m.lastActiveID = id
	fn := m.onActiveFunc
	m.mu.Unlock()
	if fn != nil {
		fn(id)
	}
}

// LastActiveID 获取最近操作的会话 ID
func (m *Manager) LastActiveID() string {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.lastActiveID
}

// normalizePath 格式化并清理路径
func normalizePath(p string) string {
	if p == "" {
		return ""
	}
	clean := filepath.Clean(p)
	// 解析可能存在的软链接以得到真实统一目录
	if realPath, err := filepath.EvalSymlinks(clean); err == nil {
		return realPath
	}
	return clean
}

func (m *Manager) GetOrCreate(id, name, shell, workDir, activeAgent, initCmd, lang string) (*Session, error) {
	return m.GetOrCreateForDir(id, name, shell, workDir, activeAgent, initCmd, lang)
}

// GetOrCreateForDir 针对项目/工作目录确保复用同一个持久会话，切换到其他项目时原会话保持在后台运行
func (m *Manager) GetOrCreateForDir(id, name, shell, workDir, activeAgent, initCmd, lang string) (*Session, error) {
	m.mu.Lock()

	targetDir := normalizePath(workDir)

	// 1. 如果指定了具体的项目/会话 ID 且存在运行中，直接复用
	if id != "" && id != "default" {
		if s, exists := m.sessions[id]; exists && s.IsRunning {
			if name != "" {
				s.mu.Lock()
				s.Name = name
				s.LastActiveAt = time.Now()
				s.mu.Unlock()
			}
			m.lastActiveID = s.ID
			fn := m.onActiveFunc
			m.mu.Unlock()
			if fn != nil {
				fn(s.ID)
			}
			return s, nil
		}
	}

	// 2. 检查是否有针对相同工作目录的正在运行的会话，若有直接复用它
	for _, s := range m.sessions {
		if s.IsRunning && normalizePath(s.WorkDir) == targetDir {
			if name != "" {
				s.mu.Lock()
				s.Name = name
				s.LastActiveAt = time.Now()
				s.mu.Unlock()
			}
			m.lastActiveID = s.ID
			fn := m.onActiveFunc
			m.mu.Unlock()
			if fn != nil {
				fn(s.ID)
			}
			return s, nil
		}
	}

	// 3. 如果没找到，确定会话 ID 并创建新会话（之前其他项目的会话继续在后台运行，绝不关闭）
	sessID := id
	if sessID == "" {
		sessID = "default"
	}

	s, err := newSession(sessID, name, shell, workDir, activeAgent, initCmd, lang)
	if err != nil {
		m.mu.Unlock()
		return nil, err
	}
	s.onActivity = m.recordActivity
	m.sessions[sessID] = s
	m.lastActiveID = sessID
	fn := m.onActiveFunc
	m.mu.Unlock()

	if fn != nil {
		fn(sessID)
	}
	return s, nil
}

// FindByProjectOrDir 查找指定项目或工作目录的当前会话（不自动新建），用于状态面板展示
func (m *Manager) FindByProjectOrDir(id, workDir string) *Session {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if id != "" {
		if s, ok := m.sessions[id]; ok {
			return s
		}
	}
	if workDir != "" {
		targetDir := normalizePath(workDir)
		for _, s := range m.sessions {
			if normalizePath(s.WorkDir) == targetDir {
				return s
			}
		}
	}
	return nil
}

func (m *Manager) Get(id string) *Session {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.sessions[id]
}

// Restart 重启指定项目的会话（仅重启目标项目，其他项目的后台会话不受影响继续运行）
func (m *Manager) Restart(id, name, shell, workDir, activeAgent, initCmd, lang string) (*Session, error) {
	m.mu.Lock()
	targetDir := normalizePath(workDir)
	for sid, s := range m.sessions {
		if sid == id || (targetDir != "" && normalizePath(s.WorkDir) == targetDir) {
			_ = s.Close()
			delete(m.sessions, sid)
		}
	}
	m.mu.Unlock()

	return m.GetOrCreateForDir(id, name, shell, workDir, activeAgent, initCmd, lang)
}

// CloseSession 停止并移除指定项目会话
func (m *Manager) CloseSession(id, workDir string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	targetDir := normalizePath(workDir)
	for sid, s := range m.sessions {
		if (id != "" && sid == id) || (targetDir != "" && normalizePath(s.WorkDir) == targetDir) {
			_ = s.Close()
			delete(m.sessions, sid)
		}
	}
}

func (m *Manager) List() []*Session {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var list []*Session
	for _, s := range m.sessions {
		list = append(list, s)
	}
	return list
}

func (m *Manager) CloseAll() {
	m.mu.Lock()
	defer m.mu.Unlock()

	for _, s := range m.sessions {
		_ = s.Close()
	}
}
