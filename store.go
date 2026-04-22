package gova

import "sync"

// StoreKey identifies a typed store. Define at package level.
// The Default value is used when no ancestor has Provided this store.
type StoreKey[T any] struct {
	Default T
	id      *storeKeyBase
	once    sync.Once
}

// storeKeyBase is the identity token: each StoreKey gets a unique one.
// Carries a dummy field so distinct allocations get distinct addresses
// (Go may coalesce addresses of zero-sized values).
type storeKeyBase struct{ _ byte }

func newStoreKeyID() *storeKeyBase {
	return &storeKeyBase{}
}

func (k *StoreKey[T]) identity() *storeKeyBase {
	k.once.Do(func() {
		k.id = newStoreKeyID()
	})
	return k.id
}

// Provide makes a store available to all descendant components in this scope's subtree.
func Provide[T any](s *Scope, key *StoreKey[T], initial T) {
	id := key.identity()
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.stores[id]; exists {
		return
	}
	s.stores[id] = newState(initial, s.onStateChange)
}

// UseStore retrieves a store from the nearest ancestor that Provided it.
// Returns a *StateValue[T] that can be read and written. Changes trigger
// re-renders in the providing scope's subtree.
// Falls back to the StoreKey's Default if no ancestor provided it.
func UseStore[T any](s *Scope, key *StoreKey[T]) *StateValue[T] {
	id := key.identity()
	val, ok := s.lookupStore(id)
	if ok {
		return val.(*StateValue[T])
	}

	// No ancestor provided: create with default at this scope
	st := newState(key.Default, s.onStateChange)
	s.mu.Lock()
	s.stores[id] = st
	s.mu.Unlock()
	return st
}
