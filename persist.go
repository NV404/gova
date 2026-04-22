package gova

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sync"
)

// PersistedState behaves like State, but the value round-trips through a file
// on disk whenever it changes. On first render, if the file already contains a
// serialized value, it seeds the state. Use this for small bits of UI state
// that should survive a `gova dev` reload, like selected tab or form input.
//
// The storage directory is picked in this order:
//
//  1. opts.Dir, if non-empty.
//  2. $GOVA_DEV_STATE (set by `gova dev`).
//  3. $XDG_STATE_HOME/gova, or ~/.cache/gova as fallback.
//
// PersistedState is explicitly NOT a general-purpose persistence layer. It
// uses JSON, writes on every change (best-effort), and is safe for small
// values only.
func PersistedState[T any](s *Scope, key string, initial T, opts ...PersistOption) *StateValue[T] {
	o := persistOptions{}
	for _, opt := range opts {
		opt(&o)
	}
	dir := resolvePersistDir(o.Dir)
	path := filepath.Join(dir, sanitizeKey(key)+".json")

	// Seed from disk if possible.
	if data, err := os.ReadFile(path); err == nil {
		var seeded T
		if jerr := json.Unmarshal(data, &seeded); jerr == nil {
			initial = seeded
		}
	}
	st := StateKey(s, "persisted:"+key, initial)

	// Only register the disk-write subscription once per (scope, key).
	// Without this gate, every render re-subscribed and a long-lived app
	// would fan out N callbacks per Set after N renders.
	subFlag := "persisted_sub:" + key
	s.mu.Lock()
	_, alreadySubscribed := s.states[subFlag]
	if !alreadySubscribed {
		s.states[subFlag] = true
	}
	s.mu.Unlock()

	if !alreadySubscribed {
		st.onChangeInternal(func(v T) { writePersisted(path, v) })
	}
	return st
}

// PersistOption configures PersistedState.
type PersistOption func(*persistOptions)

type persistOptions struct{ Dir string }

// PersistDir overrides the directory where values are written.
func PersistDir(dir string) PersistOption {
	return func(o *persistOptions) { o.Dir = dir }
}

var (
	persistMu sync.Mutex
)

func writePersisted(path string, v any) {
	persistMu.Lock()
	defer persistMu.Unlock()
	data, err := json.Marshal(v)
	if err != nil {
		return
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return
	}
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, data, 0o644); err != nil {
		return
	}
	_ = os.Rename(tmp, path)
}

func resolvePersistDir(explicit string) string {
	if explicit != "" {
		return explicit
	}
	if d := os.Getenv("GOVA_DEV_STATE"); d != "" {
		return d
	}
	if d := os.Getenv("XDG_STATE_HOME"); d != "" {
		return filepath.Join(d, "gova")
	}
	if home, err := os.UserHomeDir(); err == nil {
		return filepath.Join(home, ".cache", "gova")
	}
	return filepath.Join(os.TempDir(), "gova")
}

// sanitizeKey replaces path-unsafe characters so a user-supplied key maps to a
// single file.
func sanitizeKey(k string) string {
	out := make([]byte, 0, len(k))
	for i := 0; i < len(k); i++ {
		c := k[i]
		switch {
		case c >= 'a' && c <= 'z',
			c >= 'A' && c <= 'Z',
			c >= '0' && c <= '9',
			c == '-' || c == '_':
			out = append(out, c)
		default:
			out = append(out, '_')
		}
	}
	if len(out) == 0 {
		return "_"
	}
	return string(out)
}
