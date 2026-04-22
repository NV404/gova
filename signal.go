package gova

import (
	"fmt"
	"sync"
)

// Signal is a reactive value that notifies subscribers when it changes.
// Implemented by derivedSignal and formatSignal.
type Signal[T any] interface {
	Value() T
	subscribe(fn func(T)) (unsubscribe func())
}

// textContent is the interface for Text widget content.
// Satisfied by string (static) and Signal[string] (reactive).
type textContent interface {
	isTextContent()
}

// staticText wraps a plain string as textContent.
type staticText string

func (staticText) isTextContent() {}

// reactiveText wraps a Signal[string] as textContent.
type reactiveText struct {
	signal Signal[string]
}

func (reactiveText) isTextContent() {}

// signalBase provides shared subscription management.
type signalBase[T any] struct {
	mu          sync.RWMutex
	subscribers []signalSub[T]
	nextID      int
}

type signalSub[T any] struct {
	id int
	fn func(T)
}

func (s *signalBase[T]) addSubscriber(fn func(T)) func() {
	s.mu.Lock()
	id := s.nextID
	s.nextID++
	s.subscribers = append(s.subscribers, signalSub[T]{id: id, fn: fn})
	s.mu.Unlock()

	return func() {
		s.mu.Lock()
		defer s.mu.Unlock()
		for i, sub := range s.subscribers {
			if sub.id == id {
				s.subscribers = append(s.subscribers[:i], s.subscribers[i+1:]...)
				return
			}
		}
	}
}

func (s *signalBase[T]) notify(val T) {
	s.mu.RLock()
	subs := make([]signalSub[T], len(s.subscribers))
	copy(subs, s.subscribers)
	s.mu.RUnlock()

	for _, sub := range subs {
		sub.fn(val)
	}
}

// formatSignal derives a Signal[string] from a State by applying fmt.Sprintf.
type formatSignal[T any] struct {
	signalBase[string]
	format string
	source *StateValue[T]
	valueMu sync.RWMutex
	value   string
}

func newFormatSignal[T any](source *StateValue[T], format string) *formatSignal[T] {
	s := &formatSignal[T]{
		format: format,
		source: source,
		value:  fmt.Sprintf(format, source.Get()),
	}
	source.onChangeInternal(func(v T) {
		newVal := fmt.Sprintf(s.format, v)
		s.valueMu.Lock()
		s.value = newVal
		s.valueMu.Unlock()
		s.notify(newVal)
	})
	return s
}

func (s *formatSignal[T]) Value() string {
	s.valueMu.RLock()
	defer s.valueMu.RUnlock()
	return s.value
}

func (s *formatSignal[T]) subscribe(fn func(string)) func() {
	return s.addSubscriber(fn)
}

func (s *formatSignal[T]) isTextContent() {}

// derivedSignal derives a Signal[U] from a State[T] via a transformation function.
type derivedSignal[T any, U any] struct {
	signalBase[U]
	fn      func(T) U
	valueMu sync.RWMutex
	value   U
}

func newDerivedSignal[T any, U any](source *StateValue[T], fn func(T) U) *derivedSignal[T, U] {
	s := &derivedSignal[T, U]{
		fn:    fn,
		value: fn(source.Get()),
	}
	source.onChangeInternal(func(v T) {
		newVal := fn(v)
		s.valueMu.Lock()
		s.value = newVal
		s.valueMu.Unlock()
		s.notify(newVal)
	})
	return s
}

func (s *derivedSignal[T, U]) Value() U {
	s.valueMu.RLock()
	defer s.valueMu.RUnlock()
	return s.value
}

func (s *derivedSignal[T, U]) subscribe(fn func(U)) func() {
	return s.addSubscriber(fn)
}

// Derived creates a reactive signal that transforms state values.
// The signal auto-updates when the source state changes.
//
// Memoized by call site, so repeated renders from the same line reuse one
// subscription instead of leaking a fresh one every pass. The transform
// function is captured on first call; subsequent calls with a freshly
// allocated closure still reuse the original.
func Derived[T any, U any](state *StateValue[T], fn func(T) U) Signal[U] {
	key := "derived:" + callerKey(2)
	sig := state.getOrCreateSignal(key, func() any {
		return newDerivedSignal(state, fn)
	})
	if out, ok := sig.(Signal[U]); ok {
		return out
	}
	return newDerivedSignal(state, fn)
}
