package gova

func TextField(state *StateValue[string]) *viewNode {
	return &viewNode{
		kind: viewKindTextField,
		textField: &textFieldData{
			state: state,
		},
	}
}

func (n *viewNode) Placeholder(p string) *viewNode {
	if n.textField != nil {
		n.textField.placeholder = p
	}
	return n
}

func (n *viewNode) OnSubmit(fn func(string)) *viewNode {
	if n.textField != nil {
		n.textField.onSubmit = fn
	}
	return n
}

func (n *viewNode) Multiline() *viewNode {
	if n.textField != nil {
		n.textField.multiline = true
	}
	return n
}

func (n *viewNode) Password() *viewNode {
	if n.textField != nil {
		n.textField.password = true
	}
	return n
}
