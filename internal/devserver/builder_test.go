package devserver

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func writeModule(t *testing.T, dir, mainBody string) {
	t.Helper()
	if err := os.WriteFile(filepath.Join(dir, "go.mod"), []byte("module m\n\ngo 1.26\n"), 0o644); err != nil {
		t.Fatalf("go.mod: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, "main.go"), []byte(mainBody), 0o644); err != nil {
		t.Fatalf("main.go: %v", err)
	}
}

func TestBuildSuccess(t *testing.T) {
	dir := t.TempDir()
	writeModule(t, dir, "package main\n\nfunc main() { println(\"ok\") }\n")
	out := filepath.Join(dir, "bin")
	if runtime.GOOS == "windows" {
		out += ".exe"
	}
	if err := Build(context.Background(), dir, ".", out); err != nil {
		t.Fatalf("Build: %v", err)
	}
	if _, err := os.Stat(out); err != nil {
		t.Fatalf("output missing: %v", err)
	}
}

func TestBuildReportsCompileError(t *testing.T) {
	dir := t.TempDir()
	writeModule(t, dir, "package main\n\nfunc main() { nope }\n")
	out := filepath.Join(dir, "bin")
	err := Build(context.Background(), dir, ".", out)
	if err == nil {
		t.Fatal("expected build error")
	}
	var be *BuildError
	if !errors.As(err, &be) {
		t.Fatalf("expected *BuildError, got %T", err)
	}
	if !strings.Contains(be.Stderr, "undefined: nope") && !strings.Contains(be.Stderr, "nope") {
		t.Fatalf("expected compiler message in stderr, got %q", be.Stderr)
	}
}
