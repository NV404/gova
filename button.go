package gova

type ButtonStyle int

const (
	ButtonDefault ButtonStyle = iota
	ButtonPrimary
	ButtonDestructive
)

func Button(label string, handler func()) *viewNode {
	return &viewNode{
		kind: viewKindButton,
		button: &buttonData{
			label:   label,
			handler: handler,
		},
	}
}

func (n *viewNode) Style(s ButtonStyle) *viewNode {
	if n.button != nil {
		n.button.style = s
	}
	return n
}
