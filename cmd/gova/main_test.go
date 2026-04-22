package main

import (
	"bytes"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestRunUnknownCommandReturns2(t *testing.T) {
	var out, errb bytes.Buffer
	code := run([]string{"nope"}, &out, &errb)
	if code != 2 {
		t.Fatalf("exit code: got %d want 2", code)
	}
	if !strings.Contains(errb.String(), "unknown command") {
		t.Fatalf("expected diagnostic, got %q", errb.String())
	}
}

func TestRunNoArgsShowsUsage(t *testing.T) {
	var out, errb bytes.Buffer
	code := run(nil, &out, &errb)
	if code != 2 {
		t.Fatalf("exit code: got %d want 2", code)
	}
	if !strings.Contains(out.String(), "Usage:") {
		t.Fatalf("expected usage, got %q", out.String())
	}
}

func TestRunVersion(t *testing.T) {
	var out, errb bytes.Buffer
	code := run([]string{"version"}, &out, &errb)
	if code != 0 {
		t.Fatalf("exit code: got %d want 0", code)
	}
	if !strings.Contains(out.String(), "gova ") {
		t.Fatalf("expected version string, got %q", out.String())
	}
}

func TestRunHelp(t *testing.T) {
	var out, errb bytes.Buffer
	code := run([]string{"help"}, &out, &errb)
	if code != 0 {
		t.Fatalf("exit code: got %d", code)
	}
	if !strings.Contains(out.String(), "dev [pkg]") {
		t.Fatalf("expected help text, got %q", out.String())
	}
}

func TestRunBuildProducesBinary(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "go.mod"), []byte("module m\n\ngo 1.26\n"), 0o644); err != nil {
		t.Fatalf("go.mod: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, "main.go"),
		[]byte("package main\n\nfunc main() {}\n"), 0o644); err != nil {
		t.Fatalf("main.go: %v", err)
	}
	prev, _ := os.Getwd()
	t.Cleanup(func() { _ = os.Chdir(prev) })
	if err := os.Chdir(dir); err != nil {
		t.Fatalf("chdir: %v", err)
	}
	var out, errb bytes.Buffer
	code := run([]string{"build"}, &out, &errb)
	if code != 0 {
		t.Fatalf("build failed: code=%d stderr=%q", code, errb.String())
	}
	bin := filepath.Join(dir, "bin", filepath.Base(dir))
	if runtime.GOOS == "windows" {
		bin += ".exe"
	}
	if _, err := os.Stat(bin); err != nil {
		t.Fatalf("expected binary at %s: %v", bin, err)
	}
}
