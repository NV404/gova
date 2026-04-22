package gova

import (
	"fmt"
	"reflect"
	"sync"
)

// StateValue holds reactive state for a component. Use State() to create.
// State.Set() and State.Update() can be called from any goroutine.
type StateValue[T any] struct {
	mu             sync.RWMutex
	value          T
	onStateChange  func() // triggers component re-render scheduling
	internalSubs   []internalSub[T]
	internalSubsMu sync.Mutex
	nextSubID      int

	// signalCache memoizes derived signals (Format/Derived) so repeated
	// renders from the same call site reuse a single subscription instead
	// of registering a fresh one every pass.
	signalMu    sync.Mutex
	signalCache map[string]any
}

// internalSub is one subscriber in the internal notification list. An id
// lets us target a specific subscription for removal without relying on
// function-pointer equality (closures don't compare).
type internalSub[T any] struct {
	id int
	fn func(T)
}

func newState[T any](initial T, onStateChange func()) *StateValue[T] {
	return &StateValue[T]{
		value:         initial,
		onStateChange: onStateChange,
	}
}

// Get returns the current value.
func (s *StateValue[T]) Get() T {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.value
}

func (s *StateValue[T]) Set(v T) {
	s.mu.Lock()
	s.value = v
	s.mu.Unlock()

	s.notifyInternal(v)

	if s.onStateChange != nil {
		s.onStateChange()
	}
}

// SetSilent updates the value without triggering a re-render.
// Used internally when a UI widget syncs its own state back (e.g. Entry keystroke).
// The widget already displays the correct value: no reconciliation needed.
func (s *StateValue[T]) SetSilent(v T) {
	s.mu.Lock()
	s.value = v
	s.mu.Unlock()
}

// Update applies a transformation to the current value. The callback receives
// the current value (defensively copied for slices/maps) and must return the new value.
// The read-modify-write is atomic (holds the lock during the callback).
func (s *StateValue[T]) Update(fn func(T) T) {
	s.mu.Lock()
	s.value = fn(s.value)
	v := s.value
	s.mu.Unlock()

	s.notifyInternal(v)

	if s.onStateChange != nil {
		s.onStateChange()
	}
}

// Format returns a Signal[string] that auto-updates when this state changes.
// Uses fmt.Sprintf with the provided format string.
//
// Memoized by call site + format string, so repeated renders from the same
// line reuse one subscription instead of leaking a fresh one every pass.
func (s *StateValue[T]) Format(format string) Signal[string] {
	key := "fmt:" + callerKey(2) + ":" + format
	sig := s.getOrCreateSignal(key, func() any {
		return newFormatSignal(s, format)
	})
	if out, ok := sig.(Signal[string]); ok {
		return out
	}
	return newFormatSignal(s, format)
}

// Len returns the length of the value if it's a slice. Panics otherwise.
// Convenience for len(state.Get()) that doesn't require a full Get().
func (s *StateValue[T]) Len() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	// Use fmt to get the length via reflection: simpler than reflect package
	return lengthOf(s.value)
}

// onChangeInternal registers an internal subscriber (used by signals and
// PersistedState). Returns an unsubscribe function so callers that manage
// their own lifetime can remove themselves from the notification list.
func (s *StateValue[T]) onChangeInternal(fn func(T)) func() {
	s.internalSubsMu.Lock()
	id := s.nextSubID
	s.nextSubID++
	s.internalSubs = append(s.internalSubs, internalSub[T]{id: id, fn: fn})
	s.internalSubsMu.Unlock()

	return func() {
		s.internalSubsMu.Lock()
		defer s.internalSubsMu.Unlock()
		for i, sub := range s.internalSubs {
			if sub.id == id {
				s.internalSubs = append(s.internalSubs[:i], s.internalSubs[i+1:]...)
				return
			}
		}
	}
}

func (s *StateValue[T]) notifyInternal(v T) {
	s.internalSubsMu.Lock()
	subs := make([]internalSub[T], len(s.internalSubs))
	copy(subs, s.internalSubs)
	s.internalSubsMu.Unlock()

	for _, sub := range subs {
		sub.fn(v)
	}
}

// getOrCreateSignal looks up a derived signal in the cache, creating it
// with `create` on miss. Memoization is keyed by a caller-supplied string
// (typically the user's call site) so multiple renders from one line hand
// back the same signal instance instead of re-subscribing.
func (s *StateValue[T]) getOrCreateSignal(key string, create func() any) any {
	s.signalMu.Lock()
	defer s.signalMu.Unlock()
	if s.signalCache != nil {
		if existing, ok := s.signalCache[key]; ok {
			return existing
		}
	}
	sig := create()
	if s.signalCache == nil {
		s.signalCache = make(map[string]any)
	}
	s.signalCache[key] = sig
	return sig
}

func lengthOf(v any) int {
	rv := reflect.ValueOf(v)
	switch rv.Kind() {
	case reflect.Slice, reflect.Map, reflect.String, reflect.Array, reflect.Chan:
		return rv.Len()
	default:
		panic(fmt.Sprintf("gova: Len() called on non-collection type %T", v))
	}
}

// State creates or retrieves reactive state for the current component.
// Identity is determined automatically by the call site (file:line via runtime.Caller).
// The initial value is only used on the first render; subsequent renders return the existing state.
func State[T any](s *Scope, initial T) *StateValue[T] {
	key := callerKey(2)
	return getOrCreateState(s, key, initial)
}

// StateKey creates or retrieves reactive state with an explicit string key.
// Use this only for dynamic/conditional state where call-site identity isn't stable (rare).
func StateKey[T any](s *Scope, key string, initial T) *StateValue[T] {
	return getOrCreateState(s, "key:"+key, initial)
}

// RefValue holds a mutable value that does NOT trigger re-renders.
// Use for values that persist across renders but don't affect the view (timers, counters, refs to DOM elements).
type RefValue[T any] struct {
	mu    sync.RWMutex
	value T
}

// Get returns the current ref value.
func (r *RefValue[T]) Get() T {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.value
}

// Set updates the ref value without triggering a re-render.
func (r *RefValue[T]) Set(v T) {
	r.mu.Lock()
	r.value = v
	r.mu.Unlock()
}

// Ref creates or retrieves a non-reactive mutable reference for the current component.
// Identity is determined automatically by the call site.
func Ref[T any](s *Scope, initial T) *RefValue[T] {
	key := callerKey(2)
	return getOrCreateRef(s, key, initial)
}
