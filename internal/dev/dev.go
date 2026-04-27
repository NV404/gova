// package dev provides a convenient way to watch the working directory and restart the app
package dev

import (
	"context"
	"io/fs"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/nv404/gova/internal/utils"
)

// DevServer watches a directory tree for file changes and triggers
// a rebuild of the entry point whenever a change is detected.
type DevServer struct {
	watcher    *fsnotify.Watcher
	root       string
	entryPoint string
	debounce   time.Duration
	ctx        context.Context
	notifyCh   chan struct{}
	cmd        *exec.Cmd
	binDir     string
}

// NewDevServer creates a new DevServer rooted at root, watching all
// directories except those matching the ignore patterns. The debounce
// duration controls how long to wait after the last change before
// triggering a rebuild.
func NewDevServer(ctx context.Context, root, entryPoint string, ignore []string, debounce time.Duration) (*DevServer, error) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, err
	}

	paths, err := getWatchPaths(root, append(ignore, "gova.json"))
	if err != nil {
		watcher.Close()
		return nil, err
	}

	for _, p := range paths {
		if err := watcher.Add(p); err != nil {
			watcher.Close()
			return nil, err
		}
	}

	binDir, err := os.MkdirTemp("", "*-gova")
	if err != nil {
		return nil, err
	}

	return &DevServer{
		watcher:    watcher,
		root:       root,
		entryPoint: entryPoint,
		debounce:   debounce,
		ctx:        ctx,
		notifyCh:   make(chan struct{}, 1),
		cmd:        nil,
		binDir:     binDir,
	}, nil
}

// Start begins watching for file changes and handling rebuild triggers.
// It spawns two goroutines: one for the watcher and one for the manager.
// Cancel the context passed to NewDevServer to stop.
func (d *DevServer) Start() {
	go d.watch()
	go d.manage()
}

// watch listens for filesystem events and forwards debounced change
// notifications to the manager. Closes notifyCh when it exits so the
// manager knows to stop.
func (d *DevServer) watch() {
	defer close(d.notifyCh)

	notify := utils.Debounce(func() {
		select {
		case d.notifyCh <- struct{}{}:
		default:
		}
	}, d.debounce)

	for {
		select {
		case <-d.ctx.Done():
			return

		case event, ok := <-d.watcher.Events:
			if !ok {
				return
			}
			if event.Has(fsnotify.Create) {
				if err := d.watcher.Add(event.Name); err != nil {
					slog.Error("[dev server] failed to watch new path", "error", err, "path", event.Name)
				}
			}
			if event.Has(fsnotify.Remove) || event.Has(fsnotify.Rename) {
				d.watcher.Remove(event.Name)
			}
			notify()

		case err, ok := <-d.watcher.Errors:
			if !ok {
				return
			}
			slog.Error("[dev server] watcher error", "error", err)
		}
	}
}

// manage listens for rebuild notifications and drives the build process.
// It owns the watcher lifecycle and closes it on exit.
// It also closes the running application
func (d *DevServer) manage() {
	defer d.watcher.Close()
	defer d.killApplication()

	for {
		select {
		case <-d.ctx.Done():
			return

		case _, ok := <-d.notifyCh:
			if !ok {
				return
			}
			d.run()
		}
	}
}

// run builds the application at the binDir and runs it. Called by the manager
// whenever a debounced change notification is received.
func (d *DevServer) run() {
	slog.Info("[dev server] change detected, rebuilding...", "entry", d.entryPoint)
	d.killApplication()

	compile := exec.CommandContext(d.ctx, "go", "build", "-o", d.binDir, d.entryPoint)
	compile.Stdout = os.Stdout
	compile.Stderr = os.Stderr
	if err := compile.Run(); err != nil {
		slog.Error("[dev server] build failed", "error", err)
		return
	}

	d.cmd = exec.CommandContext(d.ctx, "/tmp/gova-bin")
	d.cmd.Stdout = os.Stdout
	d.cmd.Stderr = os.Stderr
	if err := d.cmd.Start(); err != nil {
		slog.Error("[dev server] failed to start process", "error", err)
		d.cmd = nil
		return
	}

	slog.Info("[dev server] process started", "pid", d.cmd.Process.Pid)
}

// killApplication terminates the running process (the application)
// This can be called before making a newer build and while terminating the dev server
func (d *DevServer) killApplication() {
	if d.cmd != nil && d.cmd.Process != nil {
		d.cmd.Process.Kill()
		d.cmd.Wait()
		d.cmd = nil
	}
}

// getWatchPaths walks the directory tree rooted at root and returns all
// directories that should be watched, skipping any entries that match
// the ignore patterns.
func getWatchPaths(root string, ignore []string) ([]string, error) {
	paths := make([]string, 0)
	return paths, filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		name := d.Name()
		for _, ig := range ignore {
			matched, err := filepath.Match(ig, name)
			if err != nil {
				return err
			}
			if matched {
				if d.IsDir() {
					return filepath.SkipDir
				}
				return nil
			}
		}

		if d.IsDir() {
			paths = append(paths, path)
		}

		return nil
	})
}
