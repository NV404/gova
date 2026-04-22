package devserver

import (
	"io/fs"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"
)

// Watcher recursively watches a directory tree for .go file changes and emits
// debounced notifications on Events. Close() stops the watcher and closes the
// Events channel.
type Watcher struct {
	fsw      *fsnotify.Watcher
	events   chan struct{}
	debounce time.Duration
	root     string

	mu      sync.Mutex
	closed  bool
	closeCh chan struct{}
}

// WatcherOptions configures NewWatcher.
type WatcherOptions struct {
	Root     string        // directory to watch (recursively)
	Debounce time.Duration // coalesce events within this window; defaults to 200ms
}

// NewWatcher starts a watcher on opts.Root. Events is buffered size 1; callers
// that lag will coalesce further (no missed reloads, just batched).
func NewWatcher(opts WatcherOptions) (*Watcher, error) {
	if opts.Debounce <= 0 {
		opts.Debounce = 200 * time.Millisecond
	}
	fsw, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, err
	}
	w := &Watcher{
		fsw:      fsw,
		events:   make(chan struct{}, 1),
		debounce: opts.Debounce,
		root:     opts.Root,
		closeCh:  make(chan struct{}),
	}
	if err := w.addRecursive(opts.Root); err != nil {
		fsw.Close()
		return nil, err
	}
	go w.loop()
	return w, nil
}

// Events returns a receive-only channel that fires once per debounced change.
func (w *Watcher) Events() <-chan struct{} { return w.events }

// Close stops the watcher. Safe to call multiple times.
func (w *Watcher) Close() error {
	w.mu.Lock()
	if w.closed {
		w.mu.Unlock()
		return nil
	}
	w.closed = true
	close(w.closeCh)
	w.mu.Unlock()
	return w.fsw.Close()
}

func (w *Watcher) addRecursive(root string) error {
	return filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if !d.IsDir() {
			return nil
		}
		name := d.Name()
		if name == ".git" || name == "node_modules" || name == "vendor" {
			return filepath.SkipDir
		}
		return w.fsw.Add(path)
	})
}

// shouldFire returns true for events we care about: writes/creates/removes on .go files.
// Newly created directories are added to the watch set as a side effect.
func (w *Watcher) shouldFire(ev fsnotify.Event) bool {
	if ev.Has(fsnotify.Chmod) {
		return false
	}
	if ev.Has(fsnotify.Create) {
		if info, err := statDir(ev.Name); err == nil && info.IsDir() {
			_ = w.addRecursive(ev.Name)
			return false
		}
	}
	if !strings.HasSuffix(ev.Name, ".go") {
		return false
	}
	if strings.HasSuffix(ev.Name, "_test.go") {
		return false
	}
	return true
}

func (w *Watcher) loop() {
	var timer *time.Timer
	fire := func() {
		select {
		case w.events <- struct{}{}:
		default:
			// Already pending; drop: the consumer will rebuild once
			// and see the latest tree.
		}
	}
	for {
		select {
		case <-w.closeCh:
			if timer != nil {
				timer.Stop()
			}
			close(w.events)
			return
		case ev, ok := <-w.fsw.Events:
			if !ok {
				close(w.events)
				return
			}
			if !w.shouldFire(ev) {
				continue
			}
			if timer == nil {
				timer = time.AfterFunc(w.debounce, fire)
				continue
			}
			timer.Reset(w.debounce)
		case _, ok := <-w.fsw.Errors:
			if !ok {
				close(w.events)
				return
			}
			// Non-fatal; keep watching.
		}
	}
}
