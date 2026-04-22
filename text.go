package gova

import "fmt"

// stringSignal is what Text checks for reactive string content.
type stringSignal interface {
	Value() string
	subscribe(fn func(string)) func()
}

func Text(content any) *viewNode {
	node := &viewNode{
		kind: viewKindText,
		text: &textData{},
	}

	switch v := content.(type) {
	case string:
		node.text.content = staticText(v)
	case stringSignal:
		node.text.content = reactiveText{signal: v}
	default:
		node.text.content = staticText(fmt.Sprint(content))
	}

	return node
}
