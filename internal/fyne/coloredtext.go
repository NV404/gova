package fyneBridge

import (
	"image/color"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

type coloredText struct {
	widget.BaseWidget

	text     string
	color    color.Color
	bold     bool
	italic   bool
	sizeName fyne.ThemeSizeName
	noWrap   bool
}

func newColoredText(text string, c color.Color, sizeName fyne.ThemeSizeName, bold, italic, noWrap bool) *coloredText {
	w := &coloredText{
		text:     text,
		color:    c,
		bold:     bold,
		italic:   italic,
		sizeName: sizeName,
		noWrap:   noWrap,
	}
	w.ExtendBaseWidget(w)
	return w
}

func (t *coloredText) setText(s string) {
	if t.text == s {
		return
	}
	t.text = s
	t.Refresh()
}

func (t *coloredText) setColor(c color.Color) {
	if sameColor(t.color, c) {
		return
	}
	t.color = c
	t.Refresh()
}

func (t *coloredText) CreateRenderer() fyne.WidgetRenderer {
	r := &coloredTextRenderer{owner: t}
	r.rebuild(0)
	return r
}

func sameColor(a, b color.Color) bool {
	if a == nil || b == nil {
		return a == b
	}
	ar, ag, ab, aa := a.RGBA()
	br, bg, bbl, ba := b.RGBA()
	return ar == br && ag == bg && ab == bbl && aa == ba
}

type coloredTextRenderer struct {
	owner *coloredText
	lines []*canvas.Text
	width float32
}

func (r *coloredTextRenderer) textStyle() fyne.TextStyle {
	return fyne.TextStyle{Bold: r.owner.bold, Italic: r.owner.italic}
}

func (r *coloredTextRenderer) textSize() float32 {
	th := theme.CurrentForWidget(r.owner)
	if th == nil {
		th = theme.Current()
	}
	sz := th.Size(r.owner.sizeName)
	if sz <= 0 {
		sz = th.Size(theme.SizeNameText)
	}
	return sz
}

// rebuild produces the canvas.Text instances for the current text, color,
// and width. Reuses existing canvas.Text objects when possible so the
// renderer's Objects() identity stays stable across reflows. Pass
// width=0 when the parent hasn't allocated a size yet, we'll emit a
// single line in that case.
func (r *coloredTextRenderer) rebuild(width float32) {
	r.width = width
	style := r.textStyle()
	sz := r.textSize()

	var texts []string
	if r.owner.noWrap || width <= 0 {
		texts = []string{r.owner.text}
	} else {
		texts = wrapLines(r.owner.text, width, sz, style)
	}

	for len(r.lines) < len(texts) {
		r.lines = append(r.lines, canvas.NewText("", r.owner.color))
	}
	if len(r.lines) > len(texts) {
		r.lines = r.lines[:len(texts)]
	}

	for i, t := range texts {
		ct := r.lines[i]
		ct.Text = t
		ct.TextSize = sz
		ct.TextStyle = style
		ct.Color = r.owner.color
	}
}

func (r *coloredTextRenderer) Layout(size fyne.Size) {
	r.rebuild(size.Width)

	if len(r.lines) == 1 {
		// Single-line fast path: behave like a bare canvas.Text. The
		// canvas.Text gets the full widget size, and Fyne's painter will
		// vertically center the glyph within that box (see drawText in
		// fyne.io/fyne/v2/internal/painter/gl/draw.go).
		ct := r.lines[0]
		ct.Move(fyne.NewPos(0, 0))
		ct.Resize(size)
		return
	}

	// Multi-line wrap: stack each line at a font-derived stride so
	// inter-line spacing matches what widget.Label produces. canvas.Text
	// uses its natural MinSize.Height; we don't resize it, so the painter
	// renders the glyph at its natural baseline within the line.
	style := r.textStyle()
	sz := r.textSize()
	stride := fyne.MeasureText("Mg", sz, style).Height
	for i, line := range r.lines {
		line.Move(fyne.NewPos(0, float32(i)*stride))
		line.Resize(line.MinSize())
	}
}

func (r *coloredTextRenderer) MinSize() fyne.Size {
	style := r.textStyle()
	sz := r.textSize()

	if r.owner.noWrap {
		// Same min size a bare canvas.Text would report, letting HStack
		// height fall to the natural text height keeps vertical centering
		// consistent across colored and uncolored text.
		return fyne.MeasureText(r.owner.text, sz, style)
	}

	// Wrap mode. We must allow the parent to hand us less width than
	// the full text would need, otherwise we'd never wrap. Reporting
	// the longest single word width as the minimum lets the parent
	// shrink us until that point; below it we still draw, just clipping.
	longest := float32(0)
	for _, w := range strings.Fields(r.owner.text) {
		wm := fyne.MeasureText(w, sz, style).Width
		if wm > longest {
			longest = wm
		}
	}
	if longest == 0 {
		// No whitespace; treat as a single token whose width is the text.
		longest = fyne.MeasureText(r.owner.text, sz, style).Width
	}

	if r.width <= 0 || len(r.lines) <= 1 {
		// Pre-layout, or text actually fits on one line.
		return fyne.NewSize(longest, fyne.MeasureText(r.owner.text, sz, style).Height)
	}
	stride := fyne.MeasureText("Mg", sz, style).Height
	return fyne.NewSize(longest, stride*float32(len(r.lines)))
}

func (r *coloredTextRenderer) Refresh() {
	r.rebuild(r.width)
	canvas.Refresh(r.owner)
}

func (r *coloredTextRenderer) Objects() []fyne.CanvasObject {
	objs := make([]fyne.CanvasObject, len(r.lines))
	for i, l := range r.lines {
		objs[i] = l
	}
	return objs
}

func (r *coloredTextRenderer) Destroy() {}

// wrapLines greedily packs words onto lines whose rendered width fits in
// the given pixel width. At least one word per line is guaranteed so a
// too-narrow container never produces zero-width lines (we'd rather show
// horizontal overflow than collapse the layout).
func wrapLines(text string, width, size float32, style fyne.TextStyle) []string {
	words := strings.Fields(text)
	if len(words) == 0 {
		return []string{text}
	}

	spaceW := fyne.MeasureText(" ", size, style).Width
	var lines []string
	var current []string
	var currentW float32

	for _, w := range words {
		wW := fyne.MeasureText(w, size, style).Width
		next := currentW + wW
		if len(current) > 0 {
			next += spaceW
		}
		if next <= width || len(current) == 0 {
			current = append(current, w)
			currentW = next
		} else {
			lines = append(lines, strings.Join(current, " "))
			current = []string{w}
			currentW = wW
		}
	}
	if len(current) > 0 {
		lines = append(lines, strings.Join(current, " "))
	}
	return lines
}
