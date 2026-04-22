package gova

import (
	"context"
	"sync"
)

type effectEntry struct {
	cancel  context.CancelFunc
	cleanup func()
}

func UseEffect(s *Scope, fn func(ctx context.Context) func()) {
	key := callerKey(2)

	s.mu.Lock()
	if _, exists := s.states["effect:"+key]; exists {
		s.mu.Unlock()
		return
	}
	s.states["effect:"+key] = true
	s.mu.Unlock()

	ctx, cancel := context.WithCancel(s.ctx)
	cleanup := fn(ctx)

	s.mu.Lock()
	s.states["effect_entry:"+key] = &effectEntry{
		cancel:  cancel,
		cleanup: cleanup,
	}
	s.mu.Unlock()
}

type asyncState[T any] struct {
	mu      sync.RWMutex
	data    T
	err     error
	loading bool
}

// UseAsync runs an async function and returns (data, error, loading).
// Re-runs if the component re-mounts. Cancels on unmount via context.
func UseAsync[T any](s *Scope, fn func(ctx context.Context) (T, error)) (T, error, bool) {
	key := callerKey(2)

	s.mu.Lock()
	existing, exists := s.states["async:"+key]
	if exists {
		s.mu.Unlock()
		state := existing.(*asyncState[T])
		state.mu.RLock()
		defer state.mu.RUnlock()
		return state.data, state.err, state.loading
	}

	state := &asyncState[T]{loading: true}
	s.states["async:"+key] = state
	ctx, cancel := context.WithCancel(s.ctx)
	s.states["effect_entry:"+key] = &effectEntry{cancel: cancel}
	s.mu.Unlock()

	go func() {
		data, err := fn(ctx)
		if ctx.Err() != nil {
			return
		}
		state.mu.Lock()
		state.data = data
		state.err = err
		state.loading = false
		state.mu.Unlock()

		if s.onStateChange != nil {
			s.onStateChange()
		}
	}()

	var zero T
	return zero, nil, true
}
