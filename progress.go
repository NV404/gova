package gova

func ProgressView() *viewNode {
	return &viewNode{
		kind:     viewKindProgress,
		progress: &progressData{value: -1},
	}
}

func ProgressBar(value float64) *viewNode {
	return &viewNode{
		kind:     viewKindProgress,
		progress: &progressData{value: value},
	}
}
