package gova

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func waitForFile(t *testing.T, path string, d time.Duration) []byte {
	t.Helper()
	deadline := time.Now().Add(d)
	for time.Now().Before(deadline) {
		if data, err := os.ReadFile(path); err == nil {
			return data
		}
		time.Sleep(5 * time.Millisecond)
	}
	t.Fatalf("file %s never appeared within %v", path, d)
	return nil
}

func TestPersistedStateSeedsFromDisk(t *testing.T) {
	dir := t.TempDir()
	// Pre-populate the file.
	path := filepath.Join(dir, "counter.json")
	if err := os.WriteFile(path, []byte("42"), 0o644); err != nil {
		t.Fatalf("seed: %v", err)
	}
	s := newScope(context.Background(), nil)
	defer s.destroy()
	st := PersistedState(s, "counter", 0, PersistDir(dir))
	if got := st.Get(); got != 42 {
		t.Fatalf("expected seeded value 42, got %d", got)
	}
}

func TestPersistedStateWritesOnChange(t *testing.T) {
	dir := t.TempDir()
	s := newScope(context.Background(), nil)
	defer s.destroy()
	st := PersistedState(s, "counter", 0, PersistDir(dir))
	st.Set(7)

	data := waitForFile(t, filepath.Join(dir, "counter.json"), time.Second)
	var v int
	if err := json.Unmarshal(data, &v); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if v != 7 {
		t.Fatalf("expected persisted 7, got %d", v)
	}
}

func TestPersistedStateSurvivesRestart(t *testing.T) {
	dir := t.TempDir()
	s1 := newScope(context.Background(), nil)
	st1 := PersistedState(s1, "note", "initial", PersistDir(dir))
	st1.Set("hello")
	waitForFile(t, filepath.Join(dir, "note.json"), time.Second)
	s1.destroy()

	// Simulated restart.
	s2 := newScope(context.Background(), nil)
	defer s2.destroy()
	st2 := PersistedState(s2, "note", "initial", PersistDir(dir))
	if got := st2.Get(); got != "hello" {
		t.Fatalf("expected restored %q, got %q", "hello", got)
	}
}

func TestResolvePersistDirHonorsExplicit(t *testing.T) {
	if got := resolvePersistDir("/tmp/explicit"); got != "/tmp/explicit" {
		t.Fatalf("explicit path ignored: %q", got)
	}
}

func TestResolvePersistDirReadsEnv(t *testing.T) {
	t.Setenv("GOVA_DEV_STATE", "/tmp/dev-env")
	if got := resolvePersistDir(""); got != "/tmp/dev-env" {
		t.Fatalf("expected env-driven path, got %q", got)
	}
}

func TestSanitizeKey(t *testing.T) {
	cases := map[string]string{
		"plain":        "plain",
		"with space":   "with_space",
		"slash/in/key": "slash_in_key",
		"..":           "__",
		"":             "_",
	}
	for in, want := range cases {
		if got := sanitizeKey(in); got != want {
			t.Errorf("sanitizeKey(%q): got %q want %q", in, got, want)
		}
	}
}
