package gova

import "fmt"

type componentNode struct {
	node        viewNode
	renderFn    func(*Scope) View
	scope       *Scope
	rendered    *viewNode
	errFallback func(error) View
}

// renderComponent invokes the component's renderFn. If an ErrorBoundary
// has registered a fallback on this component, panics are recovered and
// the fallback view is returned in place of the panicking subtree.
func renderComponent(c *componentNode, s *Scope) (result View) {
	if c.errFallback == nil {
		return c.renderFn(s)
	}
	defer func() {
		if r := recover(); r != nil {
			result = c.errFallback(toError(r))
		}
	}()
	return c.renderFn(s)
}

func (c *componentNode) viewNode() *viewNode {
	c.node.componentRef = c
	return &c.node
}

func Define(fn func(*Scope) View) View {
	return &componentNode{
		node:     viewNode{kind: viewKindComponent},
		renderFn: fn,
	}
}

type Viewable interface {
	Body(s *Scope) View
}

// Component wraps a Viewable as a View so it can be passed to APIs that take
// a typed View (Run, RunWithConfig, TestRender). Layout constructors accept
// a Viewable directly via their variadic any arguments; this helper is for
// the top-level entry points that cannot.
func Component(v Viewable) View {
	return wrapViewable(v)
}

// wrapViewable turns a Viewable into a View by creating a componentNode whose renderFn invokes Body with the current scope.
func wrapViewable(v Viewable) View {
	return &componentNode{
		node:     viewNode{kind: viewKindComponent},
		renderFn: func(s *Scope) View { return v.Body(s) },
	}
}

// asView converts an argument passed to a layout constructor into a View.
// Accepts View (returned as-is), Viewable (wrapped), and nil (returned
// as nil so callers can filter). Any other type panics with a message
// listing the accepted types.
func asView(x any) View {
	if x == nil {
		return nil
	}
	switch v := x.(type) {
	case View:
		return v
	case Viewable:
		return wrapViewable(v)
	default:
		panic(fmt.Sprintf("gova: cannot use %T as a View; pass a View or a Viewable", x))
	}
}
