package devserver

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"runtime"
	"sync"
	"testing"
	"time"
)

// buildHelper compiles a short Go program and returns the binary path. The
// binary sleeps until SIGTERM and writes "alive\n" then "bye\n".
func buildHelper(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	src := `package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	fmt.Println("alive:" + os.Getenv("TAG"))
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGTERM, syscall.SIGINT)
	<-ch
	fmt.Println("bye")
}
`
	writeModule(t, dir, src)
	bin := filepath.Join(dir, "app")
	if runtime.GOOS == "windows" {
		bin += ".exe"
	}
	if err := Build(context.Background(), dir, ".", bin); err != nil {
		t.Fatalf("helper build: %v", err)
	}
	return bin
}

// safeBuf is a bytes.Buffer guarded by a mutex so concurrent reads/writes
// from the test and the supervised child do not race.
type safeBuf struct {
	mu  sync.Mutex
	buf bytes.Buffer
}

func (b *safeBuf) Write(p []byte) (int, error) {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.buf.Write(p)
}

func (b *safeBuf) String() string {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.buf.String()
}

func TestSupervisorRestartSpawnsFreshChild(t *testing.T) {
	bin := buildHelper(t)

	out := &safeBuf{}
	sup := &Supervisor{
		Binary: bin,
		Stdout: out,
		Stderr: out,
		Env:    append(os.Environ(), "TAG=1"),
	}
	if err := sup.Restart(); err != nil {
		t.Fatalf("first start: %v", err)
	}
	waitFor(t, func() bool { return contains(out.String(), "alive:1") }, time.Second)

	sup.Env = append(os.Environ(), "TAG=2")
	if err := sup.Restart(); err != nil {
		t.Fatalf("second start: %v", err)
	}
	waitFor(t, func() bool { return contains(out.String(), "alive:2") }, time.Second)

	sup.Stop()
	if sup.Running() {
		t.Fatal("expected supervisor to be stopped")
	}
}

func TestSupervisorStopIsIdempotent(t *testing.T) {
	bin := buildHelper(t)
	sup := &Supervisor{Binary: bin, Stdout: &safeBuf{}, Stderr: &safeBuf{}}
	if err := sup.Restart(); err != nil {
		t.Fatalf("start: %v", err)
	}
	sup.Stop()
	sup.Stop() // must not panic or deadlock
}

func TestSupervisorStopKillsUnresponsiveChild(t *testing.T) {
	// Helper that ignores SIGTERM.
	dir := t.TempDir()
	src := `package main

import (
	"os/signal"
	"syscall"
	"time"
)

func main() {
	ch := make(chan struct{})
	signal.Ignore(syscall.SIGTERM)
	_ = ch
	time.Sleep(time.Minute)
}
`
	writeModule(t, dir, src)
	bin := filepath.Join(dir, "app")
	if runtime.GOOS == "windows" {
		bin += ".exe"
	}
	if err := Build(context.Background(), dir, ".", bin); err != nil {
		t.Fatalf("build: %v", err)
	}
	sup := &Supervisor{
		Binary:      bin,
		Stdout:      &safeBuf{},
		Stderr:      &safeBuf{},
		StopTimeout: 200 * time.Millisecond,
	}
	if err := sup.Restart(); err != nil {
		t.Fatalf("start: %v", err)
	}
	start := time.Now()
	sup.Stop()
	elapsed := time.Since(start)
	if elapsed > 2*time.Second {
		t.Fatalf("stop took too long: %v", elapsed)
	}
}

func waitFor(t *testing.T, cond func() bool, d time.Duration) {
	t.Helper()
	deadline := time.Now().Add(d)
	for time.Now().Before(deadline) {
		if cond() {
			return
		}
		time.Sleep(10 * time.Millisecond)
	}
	t.Fatalf("condition not met within %v", d)
}

func contains(haystack, needle string) bool {
	for i := 0; i+len(needle) <= len(haystack); i++ {
		if haystack[i:i+len(needle)] == needle {
			return true
		}
	}
	return false
}
