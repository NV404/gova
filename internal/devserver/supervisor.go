package devserver

import (
	"errors"
	"io"
	"os"
	"os/exec"
	"sync"
	"syscall"
	"time"
)

// Supervisor manages a single long-running child process, replacing it on
// demand. Restart is safe to call from any goroutine.
type Supervisor struct {
	// Binary is the executable path to run on each Start/Restart.
	Binary string
	// Args are passed after the binary.
	Args []string
	// Env is the child's environment. Nil inherits from the parent.
	Env []string
	// WorkDir sets the child's working directory.
	WorkDir string
	// Stdout and Stderr are the pipes the child writes to. Defaults to the
	// parent's os.Stdout/os.Stderr when nil.
	Stdout io.Writer
	Stderr io.Writer
	// StopTimeout is how long to wait after SIGTERM before SIGKILL.
	// Defaults to 2s.
	StopTimeout time.Duration

	mu  sync.Mutex
	cmd *exec.Cmd
}

// Restart stops the current child (if any) and starts a fresh one. Returns
// the start error if the new process cannot be spawned.
func (s *Supervisor) Restart() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.stopLocked()
	return s.startLocked()
}

// Stop terminates the child and does not restart it.
func (s *Supervisor) Stop() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.stopLocked()
}

// Running reports whether a child process is currently alive.
func (s *Supervisor) Running() bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.cmd != nil && s.cmd.Process != nil
}

func (s *Supervisor) startLocked() error {
	if s.Binary == "" {
		return errors.New("supervisor: Binary not set")
	}
	cmd := exec.Command(s.Binary, s.Args...)
	cmd.Dir = s.WorkDir
	if s.Env != nil {
		cmd.Env = s.Env
	}
	if s.Stdout != nil {
		cmd.Stdout = s.Stdout
	} else {
		cmd.Stdout = os.Stdout
	}
	if s.Stderr != nil {
		cmd.Stderr = s.Stderr
	} else {
		cmd.Stderr = os.Stderr
	}
	if err := cmd.Start(); err != nil {
		return err
	}
	s.cmd = cmd
	go func(c *exec.Cmd) {
		_ = c.Wait()
	}(cmd)
	return nil
}

func (s *Supervisor) stopLocked() {
	if s.cmd == nil || s.cmd.Process == nil {
		return
	}
	timeout := s.StopTimeout
	if timeout <= 0 {
		timeout = 2 * time.Second
	}
	_ = s.cmd.Process.Signal(syscall.SIGTERM)

	done := make(chan struct{})
	cmd := s.cmd
	go func() {
		_, _ = cmd.Process.Wait()
		close(done)
	}()
	select {
	case <-done:
	case <-time.After(timeout):
		_ = cmd.Process.Kill()
		<-done
	}
	s.cmd = nil
}
