package gova

import "fmt"

// ErrorBoundary wraps a child view and catches panics during rendering.
// Accepts either a View, a Viewable, or a *componentNode. Viewable values
// are wrapped into a component so their Body() call is guarded.
func ErrorBoundary(child any, fallback func(err error) View) *viewNode {
	view := asView(child)
	if view == nil {
		return &viewNode{kind: viewKindGroup}
	}
	if comp, ok := view.(*componentNode); ok {
		return safeRenderComponent(comp, fallback)
	}
	return view.viewNode()
}

func safeRenderComponent(comp *componentNode, fallback func(err error) View) *viewNode {
	origRenderFn := comp.renderFn
	comp.renderFn = func(s *Scope) View {
		var result View
		func() {
			defer func() {
				if r := recover(); r != nil {
					result = fallback(toError(r))
				}
			}()
			result = origRenderFn(s)
		}()
		return result
	}
	return comp.viewNode()
}

func toError(r any) error {
	switch v := r.(type) {
	case error:
		return v
	default:
		return fmt.Errorf("%v", v)
	}
}
