package fyneBridge

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

// tappable wraps a CanvasObject with a tap handler. Implements fyne.Tappable
// via BaseWidget embedding so screen readers and keyboard can reach it too.
type tappable struct {
	widget.BaseWidget
	content fyne.CanvasObject
	onTap   func()
}

func newTappable(content fyne.CanvasObject, onTap func()) *tappable {
	t := &tappable{content: content, onTap: onTap}
	t.ExtendBaseWidget(t)
	return t
}

func (t *tappable) CreateRenderer() fyne.WidgetRenderer {
	return widget.NewSimpleRenderer(container.NewStack(t.content))
}

func (t *tappable) Tapped(_ *fyne.PointEvent) {
	if t.onTap != nil {
		t.onTap()
	}
}

func (t *tappable) TappedSecondary(_ *fyne.PointEvent) {}
