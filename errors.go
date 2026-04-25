package gova

import "fmt"

// ErrorBoundary wraps a child view and catches panics during rendering.
// Accepts either a View, a Viewable, or a *componentNode. Viewable values
// are wrapped into a component so their Body() call is guarded.
//
// The fallback is registered on the component as a side channel; the
// rendering path consults it on each render (see renderComponent). This
// keeps the call idempotent — invoking ErrorBoundary on the same
// component across re-renders only updates the fallback, instead of
// nesting panic-recovery wrappers around renderFn.
func ErrorBoundary(child any, fallback func(err error) View) *viewNode {
	view := asView(child)
	if view == nil {
		return &viewNode{kind: viewKindGroup}
	}
	if comp, ok := view.(*componentNode); ok {
		comp.errFallback = fallback
		return comp.viewNode()
	}
	return view.viewNode()
}

func toError(r any) error {
	switch v := r.(type) {
	case error:
		return v
	default:
		return fmt.Errorf("%v", v)
	}
}
