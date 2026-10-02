//go:build windows

package session

import (
	"context"
	"os"

	"github.com/UserExistsError/conpty"
)

type windowsPTY struct {
	cpty *conpty.ConPty
}

func startPTY(shell, workDir string, cols, rows uint16) (PTYSession, error) {
	if shell == "" {
		shell = os.Getenv("COMSPEC")
		if shell == "" {
			shell = "powershell.exe"
		}
	}

	opts := []conpty.ConPtyOption{
		conpty.ConPtyDimensions(int(cols), int(rows)),
	}
	if workDir != "" {
		if stat, err := os.Stat(workDir); err == nil && stat.IsDir() {
			opts = append(opts, conpty.ConPtyWorkDir(workDir))
		}
	}

	cpty, err := conpty.Start(shell, opts...)
	if err != nil {
		// 降级尝试 cmd.exe
		cpty, err = conpty.Start("cmd.exe", opts...)
		if err != nil {
			return nil, err
		}
	}

	return &windowsPTY{
		cpty: cpty,
	}, nil
}

func (p *windowsPTY) Read(b []byte) (int, error) {
	return p.cpty.Read(b)
}

func (p *windowsPTY) Write(b []byte) (int, error) {
	return p.cpty.Write(b)
}

func (p *windowsPTY) Close() error {
	return p.cpty.Close()
}

func (p *windowsPTY) Resize(cols, rows uint16) error {
	return p.cpty.Resize(int(cols), int(rows))
}

func (p *windowsPTY) Wait() error {
	_, err := p.cpty.Wait(context.Background())
	return err
}

func (p *windowsPTY) Pid() int {
	return p.cpty.Pid()
}
