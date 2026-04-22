package gova

import (
	"fmt"
	"reflect"
	"sync"
)

// Identifiable is the interface types can implement to provide list identity.
// Used by ListOf for automatic key resolution.
type Identifiable[K comparable] interface {
	ID() K
}

// List creates a virtualized list from a state containing a slice.
// K must be comparable (used for stable identity/diffing).
func List[T any, K comparable](state *StateValue[[]T], keyFn func(T) K, renderFn func(int, T) View) *viewNode {
	items := state.Get()
	children := make([]*viewNode, len(items))
	for i, item := range items {
		child := renderFn(i, item)
		node := child.viewNode()
		node.key = keyFn(item)
		children[i] = node
	}

	return &viewNode{
		kind:     viewKindList,
		children: children,
		list: &listData{
			count: len(items),
			renderFn: func(i int) *viewNode {
				items := state.Get()
				if i >= len(items) {
					return &viewNode{kind: viewKindText, text: &textData{content: staticText("")}}
				}
				return renderFn(i, items[i]).viewNode()
			},
		},
	}
}

// ForEach iterates a plain slice. The render function may return a View
// or a Viewable struct value.
func ForEach[T any](items []T, renderFn func(T) any) *viewNode {
	children := make([]*viewNode, 0, len(items))
	for _, item := range items {
		if v := asView(renderFn(item)); v != nil {
			children = append(children, v.viewNode())
		}
	}
	return &viewNode{
		kind:     viewKindGroup,
		children: children,
		layout:   &layoutData{},
	}
}

// ForEachIndexed is ForEach but also passes the index.
func ForEachIndexed[T any](items []T, renderFn func(int, T) any) *viewNode {
	children := make([]*viewNode, 0, len(items))
	for i, item := range items {
		if v := asView(renderFn(i, item)); v != nil {
			children = append(children, v.viewNode())
		}
	}
	return &viewNode{
		kind:     viewKindGroup,
		children: children,
		layout:   &layoutData{},
	}
}

// ListOf creates a list whose key is auto-resolved from T: (1) implements
// Identifiable[K], (2) has a struct field tagged `gova:"id"`, or (3) has
// an exported `ID` field. Falls back to the slice index with a panic if
// none of the above apply (forcing the caller to use List explicitly).
func ListOf[T any](state *StateValue[[]T], renderFn func(T) any) *viewNode {
	items := state.Get()
	accessor := resolveIDAccessor[T]()
	children := make([]*viewNode, 0, len(items))
	for i, item := range items {
		v := asView(renderFn(item))
		if v == nil {
			continue
		}
		node := v.viewNode()
		if accessor != nil {
			node.key = accessor(item)
		} else {
			node.key = i
		}
		children = append(children, node)
	}
	return &viewNode{
		kind:     viewKindList,
		children: children,
		list: &listData{
			count: len(items),
			renderFn: func(i int) *viewNode {
				xs := state.Get()
				if i >= len(xs) {
					return &viewNode{kind: viewKindText, text: &textData{content: staticText("")}}
				}
				v := asView(renderFn(xs[i]))
				if v == nil {
					return &viewNode{kind: viewKindGroup}
				}
				return v.viewNode()
			},
		},
	}
}

var idAccessorCache sync.Map // reflect.Type → func(any) any

func resolveIDAccessor[T any]() func(T) any {
	var zero T
	t := reflect.TypeOf(zero)
	if t == nil {
		return nil
	}
	if cached, ok := idAccessorCache.Load(t); ok {
		if acc, ok := cached.(func(any) any); ok {
			return func(v T) any { return acc(v) }
		}
		return nil
	}
	acc := buildIDAccessor(t)
	idAccessorCache.Store(t, acc)
	if acc == nil {
		return nil
	}
	return func(v T) any { return acc(v) }
}

func buildIDAccessor(t reflect.Type) func(any) any {
	// (1) ID() method (both value and pointer receiver).
	if m, ok := t.MethodByName("ID"); ok && m.Type.NumIn() == 1 && m.Type.NumOut() == 1 {
		return func(v any) any {
			return reflect.ValueOf(v).MethodByName("ID").Call(nil)[0].Interface()
		}
	}
	// Also check pointer type for methods defined on *T.
	if pt := reflect.PointerTo(t); pt.Kind() == reflect.Ptr {
		if m, ok := pt.MethodByName("ID"); ok && m.Type.NumIn() == 1 && m.Type.NumOut() == 1 {
			return func(v any) any {
				rv := reflect.ValueOf(v)
				p := reflect.New(rv.Type())
				p.Elem().Set(rv)
				return p.MethodByName("ID").Call(nil)[0].Interface()
			}
		}
	}
	if t.Kind() != reflect.Struct {
		return nil
	}
	// (2) Struct tag `gova:"id"`.
	for i := 0; i < t.NumField(); i++ {
		f := t.Field(i)
		if f.Tag.Get("gova") == "id" && f.IsExported() {
			idx := i
			return func(v any) any { return reflect.ValueOf(v).Field(idx).Interface() }
		}
	}
	// (3) Exported field named ID.
	if f, ok := t.FieldByName("ID"); ok && f.IsExported() {
		idx := f.Index
		return func(v any) any { return reflect.ValueOf(v).FieldByIndex(idx).Interface() }
	}
	return nil
}

// ensure unused imports are bound even when package compiles alone.
var _ = fmt.Sprintf


// conditional carries both branches of a When/.Else for deferred evaluation.
type conditional struct {
	cond   bool
	ifFn   func() View
	elseFn func() View
}

// resolve renders whichever branch applies; either branch may be nil.
func (c *conditional) resolve() *viewNode {
	if c.cond && c.ifFn != nil {
		return c.ifFn().viewNode()
	}
	if !c.cond && c.elseFn != nil {
		return c.elseFn().viewNode()
	}
	return &viewNode{kind: viewKindGroup, layout: &layoutData{}}
}

// When conditionally renders a view. Returns a node that supports a chained
// .Else(fn) for the false case.
func When(cond bool, fn func() View) *viewNode {
	c := &conditional{cond: cond, ifFn: fn}
	n := c.resolve()
	n.conditional = c
	return n
}

// Else attaches a fallback branch to a When node. Evaluated only if the
// When's condition was false.
func (n *viewNode) Else(fn func() View) *viewNode {
	if n.conditional == nil {
		return n
	}
	n.conditional.elseFn = fn
	return n.conditional.resolve()
}

// Show is a shorthand for When(cond, ...) when the view is cheap/static.
// Evaluates v eagerly; use When with a func() View for lazy construction.
func Show(cond bool, v any) *viewNode {
	if cond {
		if view := asView(v); view != nil {
			return view.viewNode()
		}
	}
	return &viewNode{kind: viewKindGroup, layout: &layoutData{}}
}

// DerivedList extracts a reactive slice from a model state for use with List.
// Returns a *StateValue[[]T] that auto-updates when the source changes.
//
// Memoized by call site, so repeated renders from the same line reuse the
// same derived state instead of leaking a fresh subscription every pass.
func DerivedList[M any, T any](source *StateValue[M], extract func(M) []T) *StateValue[[]T] {
	key := "derivedlist:" + callerKey(2)
	cached := source.getOrCreateSignal(key, func() any {
		initial := extract(source.Get())
		derived := &StateValue[[]T]{value: initial}
		source.onChangeInternal(func(m M) {
			derived.mu.Lock()
			derived.value = extract(m)
			derived.mu.Unlock()
			derived.notifyInternal(derived.value)
		})
		return derived
	})
	if out, ok := cached.(*StateValue[[]T]); ok {
		return out
	}
	// Type mismatch (shouldn't happen with stable call sites) — fall back
	// to an unmemoized instance rather than panicking.
	initial := extract(source.Get())
	derived := &StateValue[[]T]{value: initial}
	source.onChangeInternal(func(m M) {
		derived.mu.Lock()
		derived.value = extract(m)
		derived.mu.Unlock()
		derived.notifyInternal(derived.value)
	})
	return derived
}
