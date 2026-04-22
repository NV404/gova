package gova

// Toggle creates a checkbox. Accepts either a *StateValue[bool] for
// state-driven binding, or a plain bool for the transient/row case.
//
//	g.Toggle(darkMode).Label("Dark Mode")
//	g.Toggle(todo.Done).OnChange(onDoneChanged)   // inside a List row
func Toggle(arg any) *viewNode {
	data := &toggleData{}
	switch v := arg.(type) {
	case *StateValue[bool]:
		data.state = v
		data.checked = v.Get()
	case bool:
		data.checked = v
	default:
		panic("gova: Toggle requires *StateValue[bool] or bool")
	}
	return &viewNode{kind: viewKindToggle, toggle: data}
}

// Label sets the toggle's label text. Works on Toggle nodes.
func (n *viewNode) Label(label string) *viewNode {
	if n.toggle != nil {
		n.toggle.label = label
	}
	return n
}

func (n *viewNode) OnChange(fn func(bool)) *viewNode {
	if n.toggle != nil {
		n.toggle.onChange = fn
	}
	return n
}
