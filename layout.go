package gova

// Alignment describes cross-axis alignment for VStack/HStack and full 2D
// alignment for ZStack. The same constants serve both; stacks ignore the
// axis that doesn't apply (VStack ignores vertical; HStack ignores horizontal).
type Alignment int

const (
	AlignCenter Alignment = iota
	AlignLeading
	AlignTrailing
	AlignTop
	AlignBottom
	AlignTopLeading
	AlignTopTrailing
	AlignBottomLeading
	AlignBottomTrailing
)

// Legacy-friendly aliases so callers can write g.Leading instead of g.AlignLeading.
const (
	Leading         = AlignLeading
	Trailing        = AlignTrailing
	Top             = AlignTop
	Bottom          = AlignBottom
	Center          = AlignCenter
	TopLeading      = AlignTopLeading
	TopTrailing     = AlignTopTrailing
	BottomLeading   = AlignBottomLeading
	BottomTrailing  = AlignBottomTrailing
)

func VStack(children ...any) *viewNode {
	return stackNode(viewKindVStack, children)
}

func HStack(children ...any) *viewNode {
	return stackNode(viewKindHStack, children)
}

// ZStack layers children on top of each other at the same position.
// Default alignment is Center; change with .Align(g.BottomTrailing) etc.
func ZStack(children ...any) *viewNode {
	return stackNode(viewKindZStack, children)
}

// Group flattens its children into the parent layout. Use to satisfy
// variadic spread limitations (Go forbids mixing spread with named args).
func Group(children ...any) *viewNode {
	return stackNode(viewKindGroup, children)
}

func stackNode(kind viewKind, children []any) *viewNode {
	node := &viewNode{
		kind:     kind,
		children: make([]*viewNode, 0, len(children)),
		layout:   &layoutData{},
	}
	for _, c := range children {
		v := asView(c)
		if v == nil {
			continue
		}
		node.children = append(node.children, v.viewNode())
	}
	return node
}

// Spacing sets inter-child gap in pixels.
func (n *viewNode) Spacing(s float32) *viewNode {
	if n.layout != nil {
		n.layout.spacing = s
	}
	return n
}

// Align sets the cross-axis alignment for VStack/HStack, or full 2D
// alignment for ZStack.
func (n *viewNode) Align(a Alignment) *viewNode {
	if n.layout != nil {
		n.layout.alignment = a
		n.layout.hasAlignment = true
	}
	return n
}

func Spacer() *viewNode {
	return &viewNode{kind: viewKindSpacer}
}

// Scaffold is the border-layout primitive. Center is always required;
// edges are optional and attached via chain methods.
//
//	g.Scaffold(g.ScrollView(content)).
//	    Top(header).
//	    Bottom(footer)
func Scaffold(center any) *viewNode {
	node := &viewNode{
		kind:     viewKindScaffold,
		scaffold: &scaffoldData{},
	}
	if v := asView(center); v != nil {
		node.children = append(node.children, v.viewNode())
	}
	return node
}

// Top attaches a view to the top edge of a Scaffold.
func (n *viewNode) Top(v any) *viewNode {
	if n.scaffold != nil {
		if view := asView(v); view != nil {
			n.scaffold.top = view
		}
	}
	return n
}

// Bottom attaches a view to the bottom edge of a Scaffold.
func (n *viewNode) Bottom(v any) *viewNode {
	if n.scaffold != nil {
		if view := asView(v); view != nil {
			n.scaffold.bottom = view
		}
	}
	return n
}

// Leading attaches a view to the leading (left in LTR) edge of a Scaffold.
func (n *viewNode) Leading(v any) *viewNode {
	if n.scaffold != nil {
		if view := asView(v); view != nil {
			n.scaffold.leading = view
		}
	}
	return n
}

// Trailing attaches a view to the trailing (right in LTR) edge of a Scaffold.
func (n *viewNode) Trailing(v any) *viewNode {
	if n.scaffold != nil {
		if view := asView(v); view != nil {
			n.scaffold.trailing = view
		}
	}
	return n
}
