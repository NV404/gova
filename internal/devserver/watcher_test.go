package devserver

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

// waitEvent blocks for an event with a deadline. Returns true if one arrived.
func waitEvent(t *testing.T, ch <-chan struct{}, d time.Duration) bool {
	t.Helper()
	select {
	case <-ch:
		return true
	case <-time.After(d):
		return false
	}
}

// warmupWatcher waits briefly for the fsnotify backend to finish registering
// its subscription. On slower CI hosts — particularly the macOS GitHub
// Actions runners — the first write can race the kqueue add and get dropped.
const watcherWarmup = 150 * time.Millisecond

// eventWait is the timeout for positive-expectation assertions. Chosen to be
// generous enough for overloaded CI without making failing tests sluggish.
const eventWait = 3 * time.Second

func TestWatcherFiresOnGoFileWrite(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "main.go"), []byte("package x"), 0o644); err != nil {
		t.Fatalf("seed: %v", err)
	}
	w, err := NewWatcher(WatcherOptions{Root: dir, Debounce: 30 * time.Millisecond})
	if err != nil {
		t.Fatalf("NewWatcher: %v", err)
	}
	defer w.Close()
	time.Sleep(watcherWarmup)

	if err := os.WriteFile(filepath.Join(dir, "main.go"), []byte("package x\n"), 0o644); err != nil {
		t.Fatalf("write: %v", err)
	}
	if !waitEvent(t, w.Events(), eventWait) {
		t.Fatal("expected debounced event after .go write")
	}
}

func TestWatcherIgnoresNonGoFiles(t *testing.T) {
	dir := t.TempDir()
	w, err := NewWatcher(WatcherOptions{Root: dir, Debounce: 30 * time.Millisecond})
	if err != nil {
		t.Fatalf("NewWatcher: %v", err)
	}
	defer w.Close()

	if err := os.WriteFile(filepath.Join(dir, "README.md"), []byte("hi"), 0o644); err != nil {
		t.Fatalf("write: %v", err)
	}
	if waitEvent(t, w.Events(), 150*time.Millisecond) {
		t.Fatal("expected no event for non-.go file")
	}
}

func TestWatcherIgnoresTestFiles(t *testing.T) {
	dir := t.TempDir()
	w, err := NewWatcher(WatcherOptions{Root: dir, Debounce: 30 * time.Millisecond})
	if err != nil {
		t.Fatalf("NewWatcher: %v", err)
	}
	defer w.Close()

	if err := os.WriteFile(filepath.Join(dir, "foo_test.go"), []byte("package x"), 0o644); err != nil {
		t.Fatalf("write: %v", err)
	}
	if waitEvent(t, w.Events(), 150*time.Millisecond) {
		t.Fatal("expected no event for _test.go file")
	}
}

func TestWatcherDebouncesBurst(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "main.go")
	if err := os.WriteFile(path, []byte("package x"), 0o644); err != nil {
		t.Fatalf("seed: %v", err)
	}
	w, err := NewWatcher(WatcherOptions{Root: dir, Debounce: 100 * time.Millisecond})
	if err != nil {
		t.Fatalf("NewWatcher: %v", err)
	}
	defer w.Close()
	time.Sleep(watcherWarmup)

	for i := 0; i < 5; i++ {
		if err := os.WriteFile(path, []byte("package x\n"), 0o644); err != nil {
			t.Fatalf("write %d: %v", i, err)
		}
		time.Sleep(10 * time.Millisecond)
	}
	if !waitEvent(t, w.Events(), eventWait) {
		t.Fatal("expected one event after burst")
	}
	// No second event should fire from the same burst.
	if waitEvent(t, w.Events(), 250*time.Millisecond) {
		t.Fatal("expected only one coalesced event")
	}
}

func TestWatcherPicksUpNewDirectory(t *testing.T) {
	dir := t.TempDir()
	w, err := NewWatcher(WatcherOptions{Root: dir, Debounce: 30 * time.Millisecond})
	if err != nil {
		t.Fatalf("NewWatcher: %v", err)
	}
	defer w.Close()
	time.Sleep(watcherWarmup)

	sub := filepath.Join(dir, "pkg")
	if err := os.Mkdir(sub, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	// Give the watcher time to register the new subdirectory, and drain any
	// directory-create signal that slipped through (it should not, since only
	// .go writes fire events, but be resilient).
	time.Sleep(watcherWarmup)
	_ = waitEvent(t, w.Events(), 50*time.Millisecond)

	if err := os.WriteFile(filepath.Join(sub, "a.go"), []byte("package pkg"), 0o644); err != nil {
		t.Fatalf("write: %v", err)
	}
	if !waitEvent(t, w.Events(), eventWait) {
		t.Fatal("expected event after writing .go in new subdirectory")
	}
}

func TestWatcherClose(t *testing.T) {
	dir := t.TempDir()
	w, err := NewWatcher(WatcherOptions{Root: dir})
	if err != nil {
		t.Fatalf("NewWatcher: %v", err)
	}
	if err := w.Close(); err != nil {
		t.Fatalf("Close: %v", err)
	}
	// Events channel must close.
	select {
	case _, ok := <-w.Events():
		if ok {
			t.Fatal("Events must be closed after Close")
		}
	case <-time.After(500 * time.Millisecond):
		t.Fatal("Events channel did not close")
	}
	// Double close must not panic or error.
	if err := w.Close(); err != nil {
		t.Fatalf("double close: %v", err)
	}
}
