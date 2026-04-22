package gova

func Grid(columns int, children ...any) *viewNode {
	node := stackNode(viewKindGrid, children)
	node.grid = &gridData{columns: columns}
	return node
}
