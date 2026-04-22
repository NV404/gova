package devserver

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// TestRunBuildsAndReloads is the end-to-end happy path: a tiny module that
// prints a version string, Run() watches it, we rewrite the version, and the
// child respawns with the new output.
func TestRunBuildsAndReloads(t *testing.T) {
	dir := t.TempDir()
	writeModule(t, dir, `package main

import (
	"fmt"
	"os/signal"
	"syscall"
)

const version = "v1"

func main() {
	fmt.Println("started:" + version)
	ch := make(chan struct{})
	go func() {
		c := make(chan any, 1)
		_ = c
		_ = signal.Ignore
		_ = syscall.SIGTERM
	}()
	// Stay alive until killed.
	sig := make(chan any)
	go func() { <-sig }()
	<-ch
}
`)
	// Replace with a version that handles SIGTERM cleanly.
	writeModule(t, dir, `package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
)

const version = "v1"

func main() {
	fmt.Println("started:" + version)
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGTERM, syscall.SIGINT)
	<-ch
}
`)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	out := &safeBuf{}
	errb := &safeBuf{}
	done := make(chan error, 1)
	go func() {
		done <- Run(ctx, Options{
			WorkDir:  dir,
			Package:  ".",
			Debounce: 80 * time.Millisecond,
			Stdout:   out,
			Stderr:   errb,
		})
	}()

	waitFor(t, func() bool { return strings.Contains(out.String(), "started:v1") }, 10*time.Second)

	// Rewrite the source; expect v2.
	src := `package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
)

const version = "v2"

func main() {
	fmt.Println("started:" + version)
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGTERM, syscall.SIGINT)
	<-ch
}
`
	if err := os.WriteFile(filepath.Join(dir, "main.go"), []byte(src), 0o644); err != nil {
		t.Fatalf("rewrite: %v", err)
	}
	waitFor(t, func() bool { return strings.Contains(out.String(), "started:v2") }, 15*time.Second)

	cancel()
	select {
	case err := <-done:
		if err != nil {
			t.Fatalf("Run returned: %v", err)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("Run did not return after cancel")
	}
}
