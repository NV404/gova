package gova

import (
	"fmt"
	"reflect"
)

// Slice helpers on *StateValue[T] when T is a slice. Methods use reflection
// to inspect the element type; the first call on a non-slice T panics with
// a clear message. Free-function counterparts (Append, Prepend, RemoveWhere,
// UpdateWhere, Replace) exist in this file for callers who prefer compile-
// time generic checks.

func (s *StateValue[T]) assertSlice(op string) reflect.Value {
	v := reflect.ValueOf(&s.value).Elem()
	if v.Kind() != reflect.Slice {
		panic(fmt.Sprintf("gova: StateValue[%T].%s requires a slice-typed state", s.value, op))
	}
	return v
}

// Append appends one or more items to a slice-typed state. Non-slice state panics.
func (s *StateValue[T]) Append(items ...any) {
	s.assertSlice("Append")
	s.mu.Lock()
	rv := reflect.ValueOf(&s.value).Elem()
	for _, item := range items {
		rv.Set(reflect.Append(rv, reflect.ValueOf(item)))
	}
	v := s.value
	s.mu.Unlock()

	s.notifyInternal(v)
	if s.onStateChange != nil {
		s.onStateChange()
	}
}

// Prepend adds an item to the front of a slice-typed state.
func (s *StateValue[T]) Prepend(item any) {
	s.assertSlice("Prepend")
	s.mu.Lock()
	rv := reflect.ValueOf(&s.value).Elem()
	one := reflect.MakeSlice(rv.Type(), 1, 1+rv.Len())
	one.Index(0).Set(reflect.ValueOf(item))
	rv.Set(reflect.AppendSlice(one, rv))
	v := s.value
	s.mu.Unlock()

	s.notifyInternal(v)
	if s.onStateChange != nil {
		s.onStateChange()
	}
}

// RemoveWhere removes all elements for which pred returns true.
// pred is called with each element as an any; callers cast inside.
func (s *StateValue[T]) RemoveWhere(pred func(item any) bool) {
	s.assertSlice("RemoveWhere")
	s.mu.Lock()
	rv := reflect.ValueOf(&s.value).Elem()
	out := reflect.MakeSlice(rv.Type(), 0, rv.Len())
	for i := 0; i < rv.Len(); i++ {
		el := rv.Index(i)
		if !pred(el.Interface()) {
			out = reflect.Append(out, el)
		}
	}
	rv.Set(out)
	v := s.value
	s.mu.Unlock()

	s.notifyInternal(v)
	if s.onStateChange != nil {
		s.onStateChange()
	}
}

// UpdateWhere applies mut to every element for which pred returns true.
// mut is called with each matching element as an any, and must return the
// replacement value (convertible to the slice element type).
func (s *StateValue[T]) UpdateWhere(pred func(item any) bool, mut func(item any) any) {
	s.assertSlice("UpdateWhere")
	s.mu.Lock()
	rv := reflect.ValueOf(&s.value).Elem()
	for i := 0; i < rv.Len(); i++ {
		el := rv.Index(i)
		if pred(el.Interface()) {
			el.Set(reflect.ValueOf(mut(el.Interface())))
		}
	}
	v := s.value
	s.mu.Unlock()

	s.notifyInternal(v)
	if s.onStateChange != nil {
		s.onStateChange()
	}
}

// Append is the type-safe free-function version of the slice method.
func Append[T any](s *StateValue[[]T], items ...T) {
	s.mu.Lock()
	s.value = append(s.value, items...)
	v := s.value
	s.mu.Unlock()

	s.notifyInternal(v)
	if s.onStateChange != nil {
		s.onStateChange()
	}
}

// Prepend is the type-safe free-function version.
func Prepend[T any](s *StateValue[[]T], item T) {
	s.mu.Lock()
	s.value = append([]T{item}, s.value...)
	v := s.value
	s.mu.Unlock()

	s.notifyInternal(v)
	if s.onStateChange != nil {
		s.onStateChange()
	}
}

// RemoveWhere is the type-safe free-function version.
func RemoveWhere[T any](s *StateValue[[]T], pred func(T) bool) {
	s.mu.Lock()
	out := s.value[:0:0]
	for _, el := range s.value {
		if !pred(el) {
			out = append(out, el)
		}
	}
	s.value = out
	v := s.value
	s.mu.Unlock()

	s.notifyInternal(v)
	if s.onStateChange != nil {
		s.onStateChange()
	}
}

// UpdateWhere is the type-safe free-function version.
func UpdateWhere[T any](s *StateValue[[]T], pred func(T) bool, mut func(T) T) {
	s.mu.Lock()
	for i, el := range s.value {
		if pred(el) {
			s.value[i] = mut(el)
		}
	}
	v := s.value
	s.mu.Unlock()

	s.notifyInternal(v)
	if s.onStateChange != nil {
		s.onStateChange()
	}
}
