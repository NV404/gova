package gova

type FillMode int

const (
	FillContain  FillMode = iota
	FillCover
	FillOriginal
	FillStretch
)

type imageData struct {
	filePath string
	url      string
	fillMode FillMode
}

func Image(filePath string) *viewNode {
	return &viewNode{
		kind: viewKindImage,
		image: &imageData{
			filePath: filePath,
			fillMode: FillContain,
		},
	}
}

func AsyncImage(url string) *viewNode {
	return &viewNode{
		kind: viewKindImage,
		image: &imageData{
			url:      url,
			fillMode: FillContain,
		},
	}
}

func (n *viewNode) Fill(mode FillMode) *viewNode {
	if n.image != nil {
		n.image.fillMode = mode
	}
	return n
}
