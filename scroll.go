package gova

func ScrollView(children ...any) *viewNode {
	node := stackNode(viewKindScrollView, children)
	node.scrollView = &scrollViewData{direction: ScrollVertical}
	return node
}

func (n *viewNode) ScrollDirection(d scrollDirection) *viewNode {
	if n.scrollView != nil {
		n.scrollView.direction = d
	}
	return n
}
