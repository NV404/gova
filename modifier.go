package gova

import "image/color"

type modifierSet struct {
	padding       *paddingModifier
	font          *FontStyle
	bold          bool
	italic        bool
	strikethrough bool
	color         any // color.Color or themeColor sentinel
	background    any // color.Color or themeColor sentinel
	disabled      bool
	opacity       float64
	hasOpacity    bool
	frame         *frameModifier
	onTap         func()
	cornerRadius  float32
	hasCorner     bool
	shadow        *shadowModifier
	stroke        *strokeModifier
	a11yLabel     string
	a11yHint      string
	a11yRole      AccessibilityRole
	hasA11y       bool
	navTitle      string
	hasNavTitle   bool
	navToolbar    *navToolbarData
	grow          bool
	noWrap        bool
}

type paddingModifier struct {
	top, right, bottom, left float32
}

type frameModifier struct {
	width, height       float32
	minWidth, minHeight float32
	maxWidth, maxHeight float32
}

type shadowModifier struct {
	color            any
	radius           float32
	offsetX, offsetY float32
}

type strokeModifier struct {
	color any
	width float32
}

// AccessibilityRole is a semantic role hint for assistive technologies.
type AccessibilityRole int

const (
	RoleNone AccessibilityRole = iota
	RoleButton
	RoleLink
	RoleImage
	RoleHeader
)

// FontStyle represents a named font style.
type FontStyle int

const (
	Body FontStyle = iota
	Title
	Title2
	Title3
	Caption
	Mono
)

// Static colors: plain RGBA, do not adapt to theme.
var (
	Red    color.Color = color.NRGBA{R: 255, G: 59, B: 48, A: 255}
	Blue   color.Color = color.NRGBA{R: 0, G: 122, B: 255, A: 255}
	Green  color.Color = color.NRGBA{R: 52, G: 199, B: 89, A: 255}
	Orange color.Color = color.NRGBA{R: 255, G: 149, B: 0, A: 255}
	Purple color.Color = color.NRGBA{R: 175, G: 82, B: 222, A: 255}
	Gray   color.Color = color.NRGBA{R: 142, G: 142, B: 147, A: 255}
	White  color.Color = color.NRGBA{R: 255, G: 255, B: 255, A: 255}
	Black  color.Color = color.NRGBA{R: 0, G: 0, B: 0, A: 255}
)

// Spacing tokens: use instead of hard-coded numbers for visual consistency.
const (
	SpaceXS float32 = 4
	SpaceSM float32 = 8
	SpaceMD float32 = 16
	SpaceLG float32 = 24
	SpaceXL float32 = 32
)

func (n *viewNode) Padding(all float32) *viewNode {
	n.modifier.padding = &paddingModifier{all, all, all, all}
	return n
}

func (n *viewNode) PaddingEdges(top, right, bottom, left float32) *viewNode {
	n.modifier.padding = &paddingModifier{top, right, bottom, left}
	return n
}

// PaddingH sets horizontal padding (leading + trailing).
func (n *viewNode) PaddingH(v float32) *viewNode {
	p := n.modifier.padding
	if p == nil {
		p = &paddingModifier{}
	}
	p.left, p.right = v, v
	n.modifier.padding = p
	return n
}

// PaddingV sets vertical padding (top + bottom).
func (n *viewNode) PaddingV(v float32) *viewNode {
	p := n.modifier.padding
	if p == nil {
		p = &paddingModifier{}
	}
	p.top, p.bottom = v, v
	n.modifier.padding = p
	return n
}

func (n *viewNode) PaddingTop(v float32) *viewNode {
	p := n.modifier.padding
	if p == nil {
		p = &paddingModifier{}
	}
	p.top = v
	n.modifier.padding = p
	return n
}

func (n *viewNode) PaddingBottom(v float32) *viewNode {
	p := n.modifier.padding
	if p == nil {
		p = &paddingModifier{}
	}
	p.bottom = v
	n.modifier.padding = p
	return n
}

func (n *viewNode) PaddingLeading(v float32) *viewNode {
	p := n.modifier.padding
	if p == nil {
		p = &paddingModifier{}
	}
	p.left = v
	n.modifier.padding = p
	return n
}

func (n *viewNode) PaddingTrailing(v float32) *viewNode {
	p := n.modifier.padding
	if p == nil {
		p = &paddingModifier{}
	}
	p.right = v
	n.modifier.padding = p
	return n
}

func (n *viewNode) Font(style FontStyle) *viewNode {
	n.modifier.font = &style
	return n
}

func (n *viewNode) Bold() *viewNode {
	n.modifier.bold = true
	return n
}

func (n *viewNode) Italic() *viewNode {
	n.modifier.italic = true
	return n
}

// Strikethrough draws a horizontal line through text. Applied via the
// Unicode combining long stroke overlay (\u0336) so it works in any
// font; has no effect on non-text nodes.
func (n *viewNode) Strikethrough(on bool) *viewNode {
	n.modifier.strikethrough = on
	return n
}

// Color sets the foreground color. Accepts a plain color.Color (fixed)
// or a semantic color constant (g.Primary, g.Accent, etc.) which resolves
// against the active theme at render time.
func (n *viewNode) Color(c any) *viewNode {
	n.modifier.color = c
	return n
}

func (n *viewNode) Background(c any) *viewNode {
	n.modifier.background = c
	return n
}

func (n *viewNode) Disabled(d bool) *viewNode {
	n.modifier.disabled = d
	return n
}

func (n *viewNode) Opacity(o float64) *viewNode {
	n.modifier.opacity = o
	n.modifier.hasOpacity = true
	return n
}

// Frame sets the preferred minimum size. Pass 0 to leave an axis unconstrained.
func (n *viewNode) Frame(width, height float32) *viewNode {
	if n.modifier.frame == nil {
		n.modifier.frame = &frameModifier{}
	}
	n.modifier.frame.width = width
	n.modifier.frame.height = height
	return n
}

// MinWidth sets the minimum width.
func (n *viewNode) MinWidth(w float32) *viewNode {
	if n.modifier.frame == nil {
		n.modifier.frame = &frameModifier{}
	}
	n.modifier.frame.minWidth = w
	return n
}

// MinHeight sets the minimum height.
func (n *viewNode) MinHeight(h float32) *viewNode {
	if n.modifier.frame == nil {
		n.modifier.frame = &frameModifier{}
	}
	n.modifier.frame.minHeight = h
	return n
}

// Grow marks a child of an HStack/VStack to expand and fill available
// space on the stack's primary axis. Siblings pack at their natural size.
// In an HStack: children before the grow child become leading, children
// after become trailing, and the grow child fills the remainder.
func (n *viewNode) Grow() *viewNode {
	n.modifier.grow = true
	return n
}

func (n *viewNode) NoWrap() *viewNode {
	n.modifier.noWrap = true
	return n
}

// OnTap makes any view tappable. Panics if applied to a Button (which
// already has a handler via its constructor).
func (n *viewNode) OnTap(fn func()) *viewNode {
	if n.kind == viewKindButton {
		panic("gova: OnTap cannot be applied to a Button; pass the handler to Button() instead")
	}
	n.modifier.onTap = fn
	return n
}

// CornerRadius rounds the view's visual box.
func (n *viewNode) CornerRadius(r float32) *viewNode {
	n.modifier.cornerRadius = r
	n.modifier.hasCorner = true
	return n
}

// Shadow casts a drop shadow from the view's visual box.
func (n *viewNode) Shadow(c any, radius, offsetX, offsetY float32) *viewNode {
	n.modifier.shadow = &shadowModifier{color: c, radius: radius, offsetX: offsetX, offsetY: offsetY}
	return n
}

// Stroke adds a colored border (stroke) around the view's visual box.
// Named Stroke (not Border) because "Border" was the name of the old
// layout primitive and would confuse readers.
func (n *viewNode) Stroke(c any, width float32) *viewNode {
	n.modifier.stroke = &strokeModifier{color: c, width: width}
	return n
}

// AccessibilityLabel sets the text read by screen readers for this view.
func (n *viewNode) AccessibilityLabel(s string) *viewNode {
	n.modifier.a11yLabel = s
	n.modifier.hasA11y = true
	return n
}

// AccessibilityHint sets additional context read by screen readers.
func (n *viewNode) AccessibilityHint(s string) *viewNode {
	n.modifier.a11yHint = s
	n.modifier.hasA11y = true
	return n
}

// AccessibilityRole hints the semantic role of this view.
func (n *viewNode) AccessibilityRole(r AccessibilityRole) *viewNode {
	n.modifier.a11yRole = r
	n.modifier.hasA11y = true
	return n
}
