package gova

import "fmt"

type componentNode struct {
	node     viewNode
	renderFn func(*Scope) View
	scope    *Scope
	rendered *viewNode
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
