package gova

import (
	"context"
	"runtime"
	"sync"
)

// Scope is the component lifecycle context. It provides access to state,
// effects, and the parent context.Context. Named "Scope" (not "Context")
// to avoid shadowing Go's context.Context.
type Scope struct {
	ctx    context.Context
	cancel context.CancelFunc

	mu     sync.Mutex
	states map[string]any
	refs   map[string]any
	stores map[any]any // keyed by *storeKeyBase, value is *StateValue[T]

	// childScopes memoizes per-component child scopes by a stable slot ID
	// so sibling components of the same type hold independent state
	// across re-renders.
	childScopes map[string]*Scope

	parent        *Scope
	onStateChange func()
}

// Context returns the underlying context.Context for cancellation and deadlines.
func (s *Scope) Context() context.Context {
	return s.ctx
}

// callerKey returns "file:line" for the caller N frames up the stack.
// This provides automatic, stable identity for State/Ref calls.
func callerKey(skip int) string {
	pc, file, line, ok := runtime.Caller(skip)
	if !ok {
		panic("gova: unable to determine caller for state identity")
	}
	// Use FuncForPC().FileLine() for stable identity across builds/inlining
	fn := runtime.FuncForPC(pc)
	if fn != nil {
		file, line = fn.FileLine(pc)
	}
	return file + ":" + itoa(line)
}

// itoa is a simple int-to-string without importing strconv.
func itoa(i int) string {
	if i == 0 {
		return "0"
	}
	if i < 0 {
		return "-" + itoa(-i)
	}
	var buf [20]byte
	pos := len(buf)
	for i > 0 {
		pos--
		buf[pos] = byte('0' + i%10)
		i /= 10
	}
	return string(buf[pos:])
}

// newScope creates a fresh scope for a component.
func newScope(parentCtx context.Context, onStateChange func()) *Scope {
	ctx, cancel := context.WithCancel(parentCtx)
	return &Scope{
		ctx:           ctx,
		cancel:        cancel,
		states:        make(map[string]any),
		refs:          make(map[string]any),
		stores:        make(map[any]any),
		onStateChange: onStateChange,
	}
}

func newChildScope(parent *Scope, onStateChange func()) *Scope {
	s := newScope(parent.ctx, onStateChange)
	s.parent = parent
	return s
}

// childScopeFor returns the child scope for a given slot ID, creating it
// on first access. Slot IDs must be stable across renders for state to
// persist: the renderer derives them from a component's Key() or its
// position within its parent's children list.
func (s *Scope) childScopeFor(id string) *Scope {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.childScopes == nil {
		s.childScopes = make(map[string]*Scope)
	}
	if cs, ok := s.childScopes[id]; ok {
		return cs
	}
	cs := newChildScope(s, s.onStateChange)
	s.childScopes[id] = cs
	return cs
}

// lookupStore walks the scope chain to find a provided store.
func (s *Scope) lookupStore(key any) (any, bool) {
	for cur := s; cur != nil; cur = cur.parent {
		cur.mu.Lock()
		val, ok := cur.stores[key]
		cur.mu.Unlock()
		if ok {
			return val, true
		}
	}
	return nil, false
}

// destroy cancels the scope's context and invokes effect cleanup functions.
// Child scopes are destroyed recursively so any effects or subscriptions
// inside nested components are torn down as well.
func (s *Scope) destroy() {
	s.mu.Lock()
	for _, v := range s.states {
		if entry, ok := v.(*effectEntry); ok {
			if entry.cleanup != nil {
				entry.cleanup()
			}
			if entry.cancel != nil {
				entry.cancel()
			}
		}
	}
	children := make([]*Scope, 0, len(s.childScopes))
	for _, cs := range s.childScopes {
		children = append(children, cs)
	}
	s.childScopes = nil
	s.mu.Unlock()

	for _, cs := range children {
		cs.destroy()
	}
	s.cancel()
}

// getOrCreateState retrieves existing state by key, or creates it with the initial value.
func getOrCreateState[T any](s *Scope, key string, initial T) *StateValue[T] {
	s.mu.Lock()
	defer s.mu.Unlock()

	if existing, ok := s.states[key]; ok {
		return existing.(*StateValue[T])
	}

	st := newState(initial, s.onStateChange)
	s.states[key] = st
	return st
}

// getOrCreateRef retrieves existing ref by key, or creates it with the initial value.
func getOrCreateRef[T any](s *Scope, key string, initial T) *RefValue[T] {
	s.mu.Lock()
	defer s.mu.Unlock()

	if existing, ok := s.refs[key]; ok {
		return existing.(*RefValue[T])
	}

	r := &RefValue[T]{value: initial}
	s.refs[key] = r
	return r
}
