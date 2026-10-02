package session

import (
	"io"
)

// PTYSession 跨平台伪终端接口抽象
type PTYSession interface {
	io.Reader
	io.Writer
	io.Closer
	Resize(cols, rows uint16) error
	Wait() error
	Pid() int
}
