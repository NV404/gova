package devserver

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"
)

// Options configures Run.
type Options struct {
	// WorkDir is the module root. Defaults to the current working directory.
	WorkDir string
	// Package is the package to build and run (e.g. ".", "./cmd/app"). Defaults to ".".
	Package string
	// Args are forwarded to the child process after the binary.
	Args []string
	// Debounce overrides the file-watch debounce window.
	Debounce time.Duration
	// StateDir is where the child should persist state across restarts. It is
	// passed to the child via GOVA_DEV_STATE. Defaults to <workdir>/.gova/dev.
	StateDir string
	// Stdout and Stderr are the streams the dev server and child write to.
	// Default to the process's os.Stdout/os.Stderr.
	Stdout io.Writer
	Stderr io.Writer
}

// Run starts the hot-reload loop and blocks until ctx is cancelled or an
// unrecoverable error is returned. A build failure is NOT unrecoverable; the
// existing child keeps running and the error is printed.
func Run(ctx context.Context, opts Options) error {
	if opts.WorkDir == "" {
		wd, err := os.Getwd()
		if err != nil {
			return err
		}
		opts.WorkDir = wd
	}
	if opts.Package == "" {
		opts.Package = "."
	}
	stdout := opts.Stdout
	if stdout == nil {
		stdout = os.Stdout
	}
	stderr := opts.Stderr
	if stderr == nil {
		stderr = os.Stderr
	}
	stateDir := opts.StateDir
	if stateDir == "" {
		stateDir = filepath.Join(opts.WorkDir, ".gova", "dev")
	}
	if err := os.MkdirAll(stateDir, 0o755); err != nil {
		return fmt.Errorf("gova dev: cannot create state dir: %w", err)
	}

	binDir, err := os.MkdirTemp("", "gova-dev-")
	if err != nil {
		return err
	}
	defer os.RemoveAll(binDir)
	binPath := filepath.Join(binDir, "app")
	if filepath.Separator == '\\' {
		binPath += ".exe"
	}

	fmt.Fprintf(stdout, "gova dev: building %s\n", opts.Package)
	if err := Build(ctx, opts.WorkDir, opts.Package, binPath); err != nil {
		fmt.Fprintln(stderr, err)
		// Keep going; user can fix and save.
	}

	sup := &Supervisor{
		Binary:  binPath,
		Args:    opts.Args,
		WorkDir: opts.WorkDir,
		Env: append(os.Environ(),
			"GOVA_DEV=1",
			"GOVA_DEV_STATE="+stateDir,
		),
		Stdout: stdout,
		Stderr: stderr,
	}
	if _, statErr := os.Stat(binPath); statErr == nil {
		if err := sup.Restart(); err != nil {
			fmt.Fprintf(stderr, "gova dev: start failed: %v\n", err)
		}
	}
	defer sup.Stop()

	watcher, err := NewWatcher(WatcherOptions{
		Root:     opts.WorkDir,
		Debounce: opts.Debounce,
	})
	if err != nil {
		return err
	}
	defer watcher.Close()

	fmt.Fprintf(stdout, "gova dev: watching %s\n", opts.WorkDir)
	for {
		select {
		case <-ctx.Done():
			return nil
		case _, ok := <-watcher.Events():
			if !ok {
				return nil
			}
			fmt.Fprintln(stdout, "gova dev: change detected, rebuilding")
			if err := Build(ctx, opts.WorkDir, opts.Package, binPath); err != nil {
				fmt.Fprintln(stderr, err)
				continue
			}
			if err := sup.Restart(); err != nil {
				fmt.Fprintf(stderr, "gova dev: restart failed: %v\n", err)
				continue
			}
			fmt.Fprintln(stdout, "gova dev: reloaded")
		}
	}
}
