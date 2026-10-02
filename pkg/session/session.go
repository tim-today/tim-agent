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
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Shell       string    `json:"shell"`
	WorkDir     string    `json:"work_dir"`
	ActiveAgent string    `json:"active_agent"`
	CreatedAt   time.Time `json:"created_at"`
	Pid         int       `json:"pid"`
	IsRunning   bool      `json:"is_running"`
	Cols        uint16    `json:"cols"`
	Rows        uint16    `json:"rows"`
	Lang        string    `json:"lang"`

	pty     PTYSession
	buffer  *RingBuffer
	clients map[*websocket.Conn]bool
	mu      sync.RWMutex
	stopCh  chan struct{}
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

	s := &Session{
		ID:          id,
		Name:        name,
		Shell:       shell,
		WorkDir:     workDir,
		ActiveAgent: activeAgent,
		CreatedAt:   time.Now(),
		Pid:         ptySess.Pid(),
		IsRunning:   true,
		Cols:        cols,
		Rows:        rows,
		Lang:        lang,
		pty:         ptySess,
		buffer:      NewRingBuffer(1024 * 1024), // 1MB 输出缓存
		clients:     make(map[*websocket.Conn]bool),
		stopCh:      make(chan struct{}),
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

// Attach 手机端或PC端连接该会话（支持断线重新连接并恢复现场）
func (s *Session) Attach(conn *websocket.Conn) {
	s.mu.Lock()
	s.clients[conn] = true
	s.mu.Unlock()

	log.Printf("[Session %s] 客户端已连接 (当前连接数: %d)", s.ID, len(s.clients))

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
			termMsg = "\r\n\x1b[31m[TimAgent] 当前会话已终止，可在控制台重启\x1b[0m\r\n"
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
		s.mu.Unlock()
		_ = conn.Close()
		log.Printf("[Session %s] 客户端已断开 (剩余连接数: %d)", s.ID, len(s.clients))
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
				_, _ = s.pty.Write(p)
			}
		} else if messageType == websocket.BinaryMessage {
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

// Manager 全局终端会话管理器
type Manager struct {
	sessions map[string]*Session
	mu       sync.RWMutex
}

var globalManager = &Manager{
	sessions: make(map[string]*Session),
}

func GetManager() *Manager {
	return globalManager
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
	m.mu.Lock()
	defer m.mu.Unlock()

	if s, exists := m.sessions[id]; exists && s.IsRunning {
		return s, nil
	}

	s, err := newSession(id, name, shell, workDir, activeAgent, initCmd, lang)
	if err != nil {
		return nil, err
	}
	m.sessions[id] = s
	return s, nil
}

// GetOrCreateForDir 针对相同工作目录确保复用同一个持久会话，实现多端/不同位置打开自动连入同一会话
func (m *Manager) GetOrCreateForDir(id, name, shell, workDir, activeAgent, initCmd, lang string) (*Session, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	targetDir := normalizePath(workDir)

	// 1. 如果指定了非默认的独立会话且存在运行中，直接复用
	if id != "" && id != "default" {
		if s, exists := m.sessions[id]; exists && s.IsRunning {
			return s, nil
		}
	}

	// 2. 检查是否有针对相同工作目录的正在运行的会话，若有直接复用它，保证无论从哪个端连入都接入同一个工作现场
	for _, s := range m.sessions {
		if s.IsRunning && normalizePath(s.WorkDir) == targetDir {
			return s, nil
		}
	}

	// 3. 如果没找到，确定会话 ID（默认用 "default" 或传入的 id）并创建新会话
	sessID := id
	if sessID == "" {
		sessID = "default"
	}

	s, err := newSession(sessID, name, shell, workDir, activeAgent, initCmd, lang)
	if err != nil {
		return nil, err
	}
	m.sessions[sessID] = s
	return s, nil
}

func (m *Manager) Get(id string) *Session {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.sessions[id]
}

func (m *Manager) Restart(id, name, shell, workDir, activeAgent, initCmd, lang string) (*Session, error) {
	m.mu.Lock()
	targetDir := normalizePath(workDir)
	// 关闭并清理旧的该 ID 或该目录下所有旧会话
	for sid, s := range m.sessions {
		if sid == id || (targetDir != "" && normalizePath(s.WorkDir) == targetDir) {
			_ = s.Close()
			delete(m.sessions, sid)
		}
	}
	m.mu.Unlock()

	return m.GetOrCreate(id, name, shell, workDir, activeAgent, initCmd, lang)
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
