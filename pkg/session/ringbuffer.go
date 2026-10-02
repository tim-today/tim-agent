package session

import "sync"

// RingBuffer 线程安全的环形缓冲区，用于保存终端最近的历史输出日志
type RingBuffer struct {
	mu       sync.RWMutex
	buf      []byte
	size     int
	start    int
	length   int
}

func NewRingBuffer(size int) *RingBuffer {
	if size <= 0 {
		size = 1024 * 512 // 默认 512KB 历史输出
	}
	return &RingBuffer{
		buf:  make([]byte, size),
		size: size,
	}
}

func (r *RingBuffer) Write(p []byte) (n int, err error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	n = len(p)
	if n == 0 {
		return 0, nil
	}

	// 如果单次写入超过缓冲区总容量，只保留最近的后半段
	if n >= r.size {
		p = p[n-r.size:]
		copy(r.buf, p)
		r.start = 0
		r.length = r.size
		return n, nil
	}

	// 环形写入
	for _, b := range p {
		pos := (r.start + r.length) % r.size
		r.buf[pos] = b
		if r.length < r.size {
			r.length++
		} else {
			r.start = (r.start + 1) % r.size
		}
	}

	return n, nil
}

// Snapshot 导出当前缓冲区中所有的历史数据副本
func (r *RingBuffer) Snapshot() []byte {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if r.length == 0 {
		return nil
	}

	result := make([]byte, r.length)
	if r.start+r.length <= r.size {
		copy(result, r.buf[r.start:r.start+r.length])
	} else {
		firstPart := r.size - r.start
		copy(result[:firstPart], r.buf[r.start:])
		copy(result[firstPart:], r.buf[:r.length-firstPart])
	}

	return result
}

// Reset 清空缓冲区
func (r *RingBuffer) Reset() {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.start = 0
	r.length = 0
}
