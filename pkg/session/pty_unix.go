//go:build !windows

package session

import (
	"os"
	"os/exec"
	"path/filepath"
	"syscall"

	"github.com/creack/pty"
)

type unixPTY struct {
	ptyFile *os.File
	cmd     *exec.Cmd
}

func startPTY(shell, workDir string, cols, rows uint16) (PTYSession, error) {
	if shell == "" {
		shell = os.Getenv("SHELL")
		if shell == "" {
			shell = "/bin/sh"
		}
	}

	var cmd *exec.Cmd
	baseShell := filepath.Base(shell)
	if baseShell == "zsh" || baseShell == "bash" {
		cmd = exec.Command(shell, "-l")
	} else {
		cmd = exec.Command(shell)
	}
	if workDir != "" {
		if stat, err := os.Stat(workDir); err == nil && stat.IsDir() {
			cmd.Dir = workDir
		}
	}

	cmd.Env = append(os.Environ(),
		"TERM=xterm-256color",
		"COLORTERM=truecolor",
		"LANG=en_US.UTF-8",
	)

	// 设置初始尺寸
	sz := &pty.Winsize{
		Rows: rows,
		Cols: cols,
	}

	ptyFile, err := pty.StartWithSize(cmd, sz)
	if err != nil {
		return nil, err
	}

	return &unixPTY{
		ptyFile: ptyFile,
		cmd:     cmd,
	}, nil
}

func (p *unixPTY) Read(b []byte) (int, error) {
	return p.ptyFile.Read(b)
}

func (p *unixPTY) Write(b []byte) (int, error) {
	return p.ptyFile.Write(b)
}

func (p *unixPTY) Close() error {
	if p.cmd != nil && p.cmd.Process != nil {
		_ = p.cmd.Process.Signal(syscall.SIGTERM)
	}
	return p.ptyFile.Close()
}

func (p *unixPTY) Resize(cols, rows uint16) error {
	sz := &pty.Winsize{
		Rows: rows,
		Cols: cols,
	}
	return pty.Setsize(p.ptyFile, sz)
}

func (p *unixPTY) Wait() error {
	if p.cmd == nil {
		return nil
	}
	return p.cmd.Wait()
}

func (p *unixPTY) Pid() int {
	if p.cmd != nil && p.cmd.Process != nil {
		return p.cmd.Process.Pid
	}
	return 0
}
