package fyneBridge

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/storage"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

type MountedNode struct {
	Spec     ViewSpec
	Object   fyne.CanvasObject
	Inner    fyne.CanvasObject
	Children []*MountedNode
	Unsubs   []func()

	FormItems []*MountedNode
	NavBar    []*MountedNode

	// Visual-modifier handles so reconcile can update colors in place
	// without remounting. Nil when the corresponding modifier wasn't set.
	BgRect      *canvas.Rectangle
	StrokeRect  *canvas.Rectangle
	FadeOverlay *canvas.Rectangle
	Tap         *tappable
}

func (m *MountedNode) innerObject() fyne.CanvasObject {
	if m.Inner != nil {
		return m.Inner
	}
	return m.Object
}

func (m *MountedNode) Cleanup() {
	for _, fn := range m.Unsubs {
		fn()
	}
	m.Unsubs = nil
	for _, child := range m.Children {
		child.Cleanup()
	}
	for _, fi := range m.FormItems {
		fi.Cleanup()
	}
	for _, nb := range m.NavBar {
		nb.Cleanup()
	}
}

type FyneBridge struct{}

func New() *FyneBridge {
	return &FyneBridge{}
}

func (b *FyneBridge) Mount(spec ViewSpec) *MountedNode {
	node := b.mountInner(spec)
	// Padding wraps first so that backgrounds/strokes enclose the padded
	// content (standard box model: padding is inside the visible box).
	if spec.HasPadding {
		padLayout := layout.NewCustomPaddedLayout(
			spec.PaddingTop, spec.PaddingBottom, spec.PaddingLeft, spec.PaddingRight,
		)
		preserveInner(node)
		node.Object = container.New(padLayout, node.Object)
	}
	b.applyVisualModifiers(node, spec)
	if spec.HasMinSize {
		spacer := canvas.NewRectangle(color.Transparent)
		spacer.SetMinSize(fyne.NewSize(spec.MinWidth, spec.MinHeight))
		preserveInner(node)
		node.Object = container.NewStack(spacer, node.Object)
	}
	return node
}

// preserveInner records the original widget as node.Inner so reconcilers
// can still reach it after wrapping. No-op if already set.
func preserveInner(node *MountedNode) {
	if node.Inner == nil {
		node.Inner = node.Object
	}
}

// applyVisualModifiers wraps the mounted object to carry background, stroke,
// opacity, and tap handler. Order: background (bottom) → content → stroke
// (top outline) → opacity overlay (top) → tap wrapper (outermost).
//
// Every wrapper exposes a handle on MountedNode so reconcileVisualModifiers
// can update colors in place without remounting.
func (b *FyneBridge) applyVisualModifiers(node *MountedNode, spec ViewSpec) {
	if spec.HasBackground {
		bg := canvas.NewRectangle(specColor(spec.BackgroundColor))
		if spec.HasCornerRadius {
			bg.CornerRadius = spec.CornerRadius
		}
		preserveInner(node)
		node.BgRect = bg
		node.Object = container.NewStack(bg, node.Object)
	}
	if spec.HasStroke {
		bg := canvas.NewRectangle(color.Transparent)
		bg.StrokeColor = specColor(spec.StrokeColor)
		bg.StrokeWidth = spec.StrokeWidth
		if spec.HasCornerRadius {
			bg.CornerRadius = spec.CornerRadius
		}
		preserveInner(node)
		node.StrokeRect = bg
		node.Object = container.NewStack(node.Object, bg)
	}
	if spec.HasOpacity {
		overlay := canvas.NewRectangle(fadeOverlayColor(spec.Opacity))
		if spec.HasCornerRadius {
			overlay.CornerRadius = spec.CornerRadius
		}
		preserveInner(node)
		node.FadeOverlay = overlay
		node.Object = container.NewStack(node.Object, overlay)
	}
	if spec.HasOnTap {
		preserveInner(node)
		tap := newTappable(node.Object, spec.OnTap)
		node.Tap = tap
		node.Object = tap
	}
}

// fadeOverlayColor returns a theme-background-colored rectangle alpha'd so
// that overlaying it on content blends toward the app background. opacity=1
// means transparent (overlay does nothing); opacity=0 means fully opaque
// background (content invisible).
func fadeOverlayColor(opacity float64) color.Color {
	if opacity < 0 {
		opacity = 0
	}
	if opacity > 1 {
		opacity = 1
	}
	base := theme.Color(theme.ColorNameBackground)
	r, g, b, _ := base.RGBA()
	return color.NRGBA{
		R: uint8(r >> 8),
		G: uint8(g >> 8),
		B: uint8(b >> 8),
		A: uint8((1 - opacity) * 255),
	}
}

// reconcileVisualModifiers updates in-place modifier handles when the spec
// changes across re-renders. If a modifier is newly added or removed, we
// can't restructure wrappers here: callers must remount.
func (b *FyneBridge) reconcileVisualModifiers(m *MountedNode, s ViewSpec) {
	if m.BgRect != nil && s.HasBackground {
		nc := specColor(s.BackgroundColor)
		if m.BgRect.FillColor != nc {
			m.BgRect.FillColor = nc
			m.BgRect.Refresh()
		}
	}
	if m.StrokeRect != nil && s.HasStroke {
		nc := specColor(s.StrokeColor)
		if m.StrokeRect.StrokeColor != nc || m.StrokeRect.StrokeWidth != s.StrokeWidth {
			m.StrokeRect.StrokeColor = nc
			m.StrokeRect.StrokeWidth = s.StrokeWidth
			m.StrokeRect.Refresh()
		}
	}
	if m.FadeOverlay != nil && s.HasOpacity {
		nc := fadeOverlayColor(s.Opacity)
		if m.FadeOverlay.FillColor != nc {
			m.FadeOverlay.FillColor = nc
			m.FadeOverlay.Refresh()
		}
	}
	if m.Tap != nil {
		m.Tap.onTap = s.OnTap
	}
}

func (b *FyneBridge) mountInner(spec ViewSpec) *MountedNode {
	switch spec.Kind {
	case KindText:
		return b.mountText(spec)
	case KindButton:
		return b.mountButton(spec)
	case KindVStack:
		return b.mountBox(spec, true)
	case KindHStack:
		return b.mountBox(spec, false)
	case KindZStack:
		return b.mountZStack(spec)
	case KindGroup:
		return b.mountBox(spec, true)
	case KindSpacer:
		return &MountedNode{Spec: spec, Object: layout.NewSpacer()}
	case KindTextField:
		return b.mountTextField(spec)
	case KindToggle:
		return b.mountToggle(spec)
	case KindSlider:
		return b.mountSlider(spec)
	case KindPicker:
		return b.mountPicker(spec)
	case KindProgress:
		return b.mountProgress(spec)
	case KindList:
		return b.mountBox(spec, true)
	case KindGrid:
		return b.mountGrid(spec)
	case KindForm:
		return b.mountForm(spec)
	case KindScrollView:
		return b.mountScroll(spec)
	case KindDivider:
		return &MountedNode{Spec: spec, Object: widget.NewSeparator()}
	case KindScaffold:
		return b.mountScaffold(spec)
	case KindTabView:
		return b.mountTabView(spec)
	case KindImage:
		return b.mountImage(spec)
	case KindNavStack:
		return b.mountNavStack(spec)
	default:
		return &MountedNode{Spec: spec, Object: widget.NewLabel("(unknown)")}
	}
}

func (b *FyneBridge) Reconcile(mounted *MountedNode, newSpec ViewSpec) {
	if mounted.Spec.Kind != newSpec.Kind {
		return
	}

	b.reconcileVisualModifiers(mounted, newSpec)

	switch newSpec.Kind {
	case KindText:
		b.reconcileText(mounted, newSpec)
	case KindButton:
		b.reconcileButton(mounted, newSpec)
	case KindTextField:
		b.reconcileTextField(mounted, newSpec)
	case KindToggle:
		b.reconcileToggle(mounted, newSpec)
	case KindSlider:
		b.reconcileSlider(mounted, newSpec)
	case KindPicker:
		b.reconcilePicker(mounted, newSpec)
	case KindProgress:
		b.reconcileProgress(mounted, newSpec)
	case KindVStack, KindHStack, KindZStack, KindGroup, KindList, KindGrid:
		b.reconcileContainer(mounted, newSpec)
	case KindForm:
		b.reconcileForm(mounted, newSpec)
	case KindScrollView:
		b.reconcileScroll(mounted, newSpec)
	case KindScaffold:
		b.reconcileScaffold(mounted, newSpec)
	case KindTabView:
		b.reconcileTabView(mounted, newSpec)
	case KindNavStack:
		b.reconcileNavStack(mounted, newSpec)
	}

	mounted.Spec = newSpec
}

// --- Mount helpers ---

func fontSizeName(fontStyle int) fyne.ThemeSizeName {
	switch fontStyle {
	case 1:
		return theme.SizeNameHeadingText
	case 2:
		return theme.SizeNameSubHeadingText
	case 3:
		return theme.SizeNameSubHeadingText
	case 4:
		return theme.SizeNameCaptionText
	default:
		return theme.SizeNameText
	}
}

func (b *FyneBridge) mountText(spec ViewSpec) *MountedNode {
	if spec.HasColor {
		ct := canvas.NewText(spec.TextValue, specColor(spec.TextColor))
		ct.TextSize = theme.Size(fontSizeName(spec.FontStyle))
		ct.TextStyle.Bold = spec.Bold
		ct.TextStyle.Italic = spec.Italic
		node := &MountedNode{Spec: spec, Object: ct}
		if spec.TextSubscribe != nil {
			unsub := spec.TextSubscribe(func(newText string) {
				fyne.Do(func() {
					ct.Text = newText
					ct.Refresh()
				})
			})
			node.Unsubs = append(node.Unsubs, unsub)
		}
		return node
	}

	if spec.FontStyle != 0 {
		seg := &widget.TextSegment{
			Text: spec.TextValue,
			Style: widget.RichTextStyle{
				SizeName:  fontSizeName(spec.FontStyle),
				TextStyle: fyne.TextStyle{Bold: spec.Bold, Italic: spec.Italic},
			},
		}
		rt := widget.NewRichText(seg)
		node := &MountedNode{Spec: spec, Object: rt}
		if spec.TextSubscribe != nil {
			unsub := spec.TextSubscribe(func(newText string) {
				fyne.Do(func() {
					seg.Text = newText
					rt.Refresh()
				})
			})
			node.Unsubs = append(node.Unsubs, unsub)
		}
		return node
	}

	label := widget.NewLabel(spec.TextValue)
	if spec.Bold {
		label.TextStyle.Bold = true
	}
	if spec.Italic {
		label.TextStyle.Italic = true
	}
	node := &MountedNode{Spec: spec, Object: label}
	if spec.TextSubscribe != nil {
		unsub := spec.TextSubscribe(func(newText string) {
			fyne.Do(func() {
				label.SetText(newText)
			})
		})
		node.Unsubs = append(node.Unsubs, unsub)
	}
	return node
}

func specColor(c [4]uint8) color.Color {
	return color.NRGBA{R: c[0], G: c[1], B: c[2], A: c[3]}
}

func (b *FyneBridge) mountButton(spec ViewSpec) *MountedNode {
	btn := widget.NewButton(spec.ButtonLabel, spec.ButtonHandler)
	return &MountedNode{Spec: spec, Object: btn}
}

func (b *FyneBridge) mountTextField(spec ViewSpec) *MountedNode {
	entry := widget.NewEntry()
	entry.PlaceHolder = spec.TextFieldPlaceholder
	if spec.TextFieldMultiline {
		entry.MultiLine = true
	}
	if spec.TextFieldPassword {
		entry.Password = true
	}
	if spec.TextFieldGet != nil {
		entry.SetText(spec.TextFieldGet())
	}

	entry.OnChanged = func(text string) {
		if spec.TextFieldSetSilent != nil {
			spec.TextFieldSetSilent(text)
		}
	}
	if spec.TextFieldOnSubmit != nil {
		entry.OnSubmitted = spec.TextFieldOnSubmit
	}

	return &MountedNode{Spec: spec, Object: entry}
}

func (b *FyneBridge) mountToggle(spec ViewSpec) *MountedNode {
	check := widget.NewCheck(spec.ToggleLabel, spec.ToggleOnChange)
	check.Checked = spec.ToggleChecked
	return &MountedNode{Spec: spec, Object: check}
}

func (b *FyneBridge) mountSlider(spec ViewSpec) *MountedNode {
	s := widget.NewSlider(spec.SliderMin, spec.SliderMax)
	s.Value = spec.SliderValue
	if spec.SliderStep > 0 {
		s.Step = spec.SliderStep
	}
	if spec.SliderOnChange != nil {
		s.OnChanged = spec.SliderOnChange
	}
	return &MountedNode{Spec: spec, Object: s}
}

func (b *FyneBridge) mountPicker(spec ViewSpec) *MountedNode {
	sel := widget.NewSelect(spec.PickerOptions, spec.PickerOnChange)
	if spec.PickerSelected >= 0 && spec.PickerSelected < len(spec.PickerOptions) {
		sel.SetSelected(spec.PickerOptions[spec.PickerSelected])
	}
	return &MountedNode{Spec: spec, Object: sel}
}

func (b *FyneBridge) mountProgress(spec ViewSpec) *MountedNode {
	if spec.ProgressValue < 0 {
		return &MountedNode{Spec: spec, Object: widget.NewProgressBarInfinite()}
	}
	bar := widget.NewProgressBar()
	bar.SetValue(spec.ProgressValue)
	return &MountedNode{Spec: spec, Object: bar}
}

func (b *FyneBridge) mountBox(spec ViewSpec, vertical bool) *MountedNode {
	children := make([]*MountedNode, len(spec.Children))
	objects := make([]fyne.CanvasObject, len(spec.Children))
	growIdx := -1
	for i, childSpec := range spec.Children {
		child := b.Mount(childSpec)
		children[i] = child
		objects[i] = child.Object
		if childSpec.Grow && growIdx == -1 {
			growIdx = i
		}
	}

	if growIdx >= 0 {
		// Use Border layout: siblings before/after the growing child pack at
		// their natural size; the grow child fills the remainder.
		var before, after fyne.CanvasObject
		if growIdx > 0 {
			if vertical {
				before = container.NewVBox(objects[:growIdx]...)
			} else {
				before = container.NewHBox(objects[:growIdx]...)
			}
		}
		if growIdx < len(objects)-1 {
			if vertical {
				after = container.NewVBox(objects[growIdx+1:]...)
			} else {
				after = container.NewHBox(objects[growIdx+1:]...)
			}
		}
		var wrap fyne.CanvasObject
		if vertical {
			wrap = container.NewBorder(before, after, nil, nil, objects[growIdx])
		} else {
			wrap = container.NewBorder(nil, nil, before, after, objects[growIdx])
		}
		return &MountedNode{Spec: spec, Object: wrap, Children: children}
	}

	var c *fyne.Container
	if vertical {
		c = container.NewVBox(objects...)
	} else {
		c = container.NewHBox(objects...)
	}

	return &MountedNode{Spec: spec, Object: c, Children: children}
}

func (b *FyneBridge) mountZStack(spec ViewSpec) *MountedNode {
	children := make([]*MountedNode, len(spec.Children))
	objects := make([]fyne.CanvasObject, len(spec.Children))
	for i, childSpec := range spec.Children {
		child := b.Mount(childSpec)
		children[i] = child
		objects[i] = child.Object
	}
	c := container.NewStack(objects...)
	return &MountedNode{Spec: spec, Object: c, Children: children}
}

func (b *FyneBridge) mountGrid(spec ViewSpec) *MountedNode {
	children := make([]*MountedNode, len(spec.Children))
	objects := make([]fyne.CanvasObject, len(spec.Children))
	for i, childSpec := range spec.Children {
		child := b.Mount(childSpec)
		children[i] = child
		objects[i] = child.Object
	}
	cols := spec.GridColumns
	if cols < 1 {
		cols = 2
	}
	c := container.New(layout.NewGridLayoutWithColumns(cols), objects...)
	return &MountedNode{Spec: spec, Object: c, Children: children}
}

func (b *FyneBridge) mountForm(spec ViewSpec) *MountedNode {
	formItems := make([]*MountedNode, len(spec.FormItems))
	fyneItems := make([]*widget.FormItem, len(spec.FormItems))
	for i, fi := range spec.FormItems {
		child := b.Mount(fi.View)
		formItems[i] = child
		fyneItems[i] = widget.NewFormItem(fi.Label, child.Object)
	}
	form := widget.NewForm(fyneItems...)
	return &MountedNode{Spec: spec, Object: form, FormItems: formItems}
}

func (b *FyneBridge) mountScaffold(spec ViewSpec) *MountedNode {
	var top, bottom, left, right fyne.CanvasObject
	node := &MountedNode{Spec: spec}

	if spec.Scaffold != nil {
		if spec.Scaffold.Top != nil {
			m := b.Mount(*spec.Scaffold.Top)
			top = m.Object
			node.Children = append(node.Children, m)
		}
		if spec.Scaffold.Bottom != nil {
			m := b.Mount(*spec.Scaffold.Bottom)
			bottom = m.Object
			node.Children = append(node.Children, m)
		}
		if spec.Scaffold.Leading != nil {
			m := b.Mount(*spec.Scaffold.Leading)
			left = m.Object
			node.Children = append(node.Children, m)
		}
		if spec.Scaffold.Trailing != nil {
			m := b.Mount(*spec.Scaffold.Trailing)
			right = m.Object
			node.Children = append(node.Children, m)
		}
	}

	centerObjects := make([]fyne.CanvasObject, 0, len(spec.Children))
	for _, childSpec := range spec.Children {
		m := b.Mount(childSpec)
		node.Children = append(node.Children, m)
		centerObjects = append(centerObjects, m.Object)
	}

	var center fyne.CanvasObject
	if len(centerObjects) == 1 {
		center = centerObjects[0]
	} else if len(centerObjects) > 1 {
		center = container.NewVBox(centerObjects...)
	}

	node.Object = container.NewBorder(top, bottom, left, right, center)
	return node
}

func (b *FyneBridge) mountNavStack(spec ViewSpec) *MountedNode {
	node := &MountedNode{Spec: spec}

	var topObj fyne.CanvasObject
	if len(spec.Children) > 0 {
		child := b.Mount(spec.Children[0])
		node.Children = append(node.Children, child)
		topObj = child.Object
	}

	barObj, barChildren := b.buildNavBar(spec.NavBar)
	node.NavBar = barChildren

	// Wrap in a stable outer container so reconcileNavStack can swap the
	// inner layout without invalidating the parent's reference to Object.
	inner := container.NewBorder(barObj, nil, nil, nil, topObj)
	node.Object = container.NewStack(inner)
	return node
}

func (b *FyneBridge) buildNavBar(bar *NavBarSpec) (fyne.CanvasObject, []*MountedNode) {
	if bar == nil {
		return nil, nil
	}
	var mounted []*MountedNode
	var leading []fyne.CanvasObject
	if bar.CanPop {
		back := widget.NewButton("< Back", bar.OnBack)
		leading = append(leading, back)
	}
	for _, s := range bar.Leading {
		m := (&FyneBridge{}).Mount(s)
		mounted = append(mounted, m)
		leading = append(leading, m.Object)
	}
	var trailing []fyne.CanvasObject
	for _, s := range bar.Trailing {
		m := (&FyneBridge{}).Mount(s)
		mounted = append(mounted, m)
		trailing = append(trailing, m.Object)
	}

	title := widget.NewLabel(bar.Title)
	title.TextStyle.Bold = true
	title.Alignment = fyne.TextAlignCenter

	leadingBox := container.NewHBox(leading...)
	trailingBox := container.NewHBox(trailing...)

	row := container.NewBorder(nil, nil, leadingBox, trailingBox, title)
	return container.NewVBox(row, widget.NewSeparator()), mounted
}

func (b *FyneBridge) mountTabView(spec ViewSpec) *MountedNode {
	children := make([]*MountedNode, len(spec.Children))
	tabs := make([]*container.TabItem, len(spec.Children))
	for i, childSpec := range spec.Children {
		child := b.Mount(childSpec)
		children[i] = child
		label := ""
		if i < len(spec.TabLabels) {
			label = spec.TabLabels[i]
		}
		tabs[i] = container.NewTabItem(label, child.Object)
	}

	appTabs := container.NewAppTabs(tabs...)
	switch spec.TabPlacement {
	case 1:
		appTabs.SetTabLocation(container.TabLocationBottom)
	case 2:
		appTabs.SetTabLocation(container.TabLocationLeading)
	case 3:
		appTabs.SetTabLocation(container.TabLocationTrailing)
	default:
		appTabs.SetTabLocation(container.TabLocationTop)
	}

	return &MountedNode{Spec: spec, Object: appTabs, Children: children}
}

func (b *FyneBridge) mountImage(spec ViewSpec) *MountedNode {
	var img *canvas.Image
	if spec.ImageFilePath != "" {
		img = canvas.NewImageFromFile(spec.ImageFilePath)
	} else if spec.ImageURL != "" {
		parsed, err := storage.ParseURI(spec.ImageURL)
		if err == nil {
			img = canvas.NewImageFromURI(parsed)
		} else {
			img = canvas.NewImageFromFile("")
		}
	} else {
		img = canvas.NewImageFromFile("")
	}

	switch spec.ImageFillMode {
	case 1:
		img.FillMode = canvas.ImageFillContain
	case 2:
		img.FillMode = canvas.ImageFillOriginal
	case 3:
		img.FillMode = canvas.ImageFillStretch
	default:
		img.FillMode = canvas.ImageFillContain
	}

	img.SetMinSize(fyne.NewSize(100, 100))
	return &MountedNode{Spec: spec, Object: img}
}

func (b *FyneBridge) mountScroll(spec ViewSpec) *MountedNode {
	inner := b.mountBox(spec, true)
	scroll := container.NewScroll(inner.Object)
	return &MountedNode{Spec: spec, Object: scroll, Children: []*MountedNode{inner}}
}

// --- Reconcile helpers ---

func (b *FyneBridge) reconcileText(m *MountedNode, s ViewSpec) {
	inner := m.innerObject()

	if ct, ok := inner.(*canvas.Text); ok {
		changed := false
		if s.TextValue != m.Spec.TextValue && s.TextSubscribe == nil {
			ct.Text = s.TextValue
			changed = true
		}
		if s.HasColor && s.TextColor != m.Spec.TextColor {
			ct.Color = specColor(s.TextColor)
			changed = true
		}
		if changed {
			ct.Refresh()
		}
		return
	}

	if rt, ok := inner.(*widget.RichText); ok {
		if s.TextValue != m.Spec.TextValue && s.TextSubscribe == nil {
			if len(rt.Segments) > 0 {
				if seg, ok := rt.Segments[0].(*widget.TextSegment); ok {
					seg.Text = s.TextValue
					rt.Refresh()
				}
			}
		}
		return
	}

	if label, ok := inner.(*widget.Label); ok {
		if s.TextValue != m.Spec.TextValue && s.TextSubscribe == nil {
			label.SetText(s.TextValue)
		}
		if s.Bold != m.Spec.Bold {
			label.TextStyle.Bold = s.Bold
			label.Refresh()
		}
	}
}

func (b *FyneBridge) reconcileButton(m *MountedNode, s ViewSpec) {
	btn, ok := m.innerObject().(*widget.Button)
	if !ok {
		return
	}
	if s.ButtonLabel != m.Spec.ButtonLabel {
		btn.SetText(s.ButtonLabel)
	}
	btn.OnTapped = s.ButtonHandler
}

func (b *FyneBridge) reconcileTextField(m *MountedNode, s ViewSpec) {
	entry, ok := m.innerObject().(*widget.Entry)
	if !ok {
		return
	}
	if s.TextFieldOnSubmit != nil {
		entry.OnSubmitted = s.TextFieldOnSubmit
	}
	entry.OnChanged = func(text string) {
		if s.TextFieldSetSilent != nil {
			s.TextFieldSetSilent(text)
		}
	}
	if s.TextFieldGet != nil {
		stateText := s.TextFieldGet()
		if entry.Text != stateText {
			entry.SetText(stateText)
		}
	}
}

func (b *FyneBridge) reconcileToggle(m *MountedNode, s ViewSpec) {
	check, ok := m.innerObject().(*widget.Check)
	if !ok {
		return
	}
	check.OnChanged = s.ToggleOnChange
	if s.ToggleChecked != check.Checked {
		check.Checked = s.ToggleChecked
		check.Refresh()
	}
}

func (b *FyneBridge) reconcileSlider(m *MountedNode, s ViewSpec) {
	sl, ok := m.innerObject().(*widget.Slider)
	if !ok {
		return
	}
	if s.SliderValue != sl.Value {
		sl.Value = s.SliderValue
		sl.Refresh()
	}
	sl.OnChanged = s.SliderOnChange
}

func (b *FyneBridge) reconcilePicker(m *MountedNode, s ViewSpec) {
	sel, ok := m.innerObject().(*widget.Select)
	if !ok {
		return
	}
	sel.OnChanged = s.PickerOnChange
	if s.PickerSelected >= 0 && s.PickerSelected < len(s.PickerOptions) {
		want := s.PickerOptions[s.PickerSelected]
		if sel.Selected != want {
			sel.SetSelected(want)
		}
	}
}

func (b *FyneBridge) reconcileProgress(m *MountedNode, s ViewSpec) {
	if bar, ok := m.innerObject().(*widget.ProgressBar); ok && s.ProgressValue >= 0 {
		bar.SetValue(s.ProgressValue)
	}
}

func (b *FyneBridge) reconcileContainer(m *MountedNode, s ViewSpec) {
	c, ok := m.innerObject().(*fyne.Container)
	if !ok {
		return
	}

	oldLen := len(m.Children)
	newLen := len(s.Children)
	minLen := oldLen
	if newLen < minLen {
		minLen = newLen
	}

	for i := 0; i < minLen; i++ {
		oldChild := m.Children[i]
		newChildSpec := s.Children[i]

		kindMatch := oldChild.Spec.Kind == newChildSpec.Kind
		keyMismatch := oldChild.Spec.Key != nil && newChildSpec.Key != nil && oldChild.Spec.Key != newChildSpec.Key

		if kindMatch && !keyMismatch {
			b.Reconcile(oldChild, newChildSpec)
		} else {
			oldChild.Cleanup()
			newChild := b.Mount(newChildSpec)
			m.Children[i] = newChild
			c.Objects[i] = newChild.Object
		}
	}

	if newLen < oldLen {
		for i := newLen; i < oldLen; i++ {
			m.Children[i].Cleanup()
		}
		m.Children = m.Children[:newLen]
		c.Objects = c.Objects[:newLen]
	}

	if newLen > oldLen {
		for i := oldLen; i < newLen; i++ {
			newChild := b.Mount(s.Children[i])
			m.Children = append(m.Children, newChild)
			c.Objects = append(c.Objects, newChild.Object)
		}
	}

	if oldLen != newLen {
		c.Refresh()
	}
}

func (b *FyneBridge) reconcileForm(m *MountedNode, s ViewSpec) {
	for i := 0; i < len(s.FormItems) && i < len(m.FormItems); i++ {
		if m.FormItems[i].Spec.Kind == s.FormItems[i].View.Kind {
			b.Reconcile(m.FormItems[i], s.FormItems[i].View)
		}
	}
}

func (b *FyneBridge) reconcileScroll(m *MountedNode, s ViewSpec) {
	if len(m.Children) > 0 {
		b.reconcileContainer(m.Children[0], s)
		if scroll, ok := m.innerObject().(*container.Scroll); ok {
			scroll.Refresh()
		}
	}
}

func (b *FyneBridge) reconcileScaffold(m *MountedNode, s ViewSpec) {
	var newChildSpecs []ViewSpec
	if s.Scaffold != nil {
		if s.Scaffold.Top != nil {
			newChildSpecs = append(newChildSpecs, *s.Scaffold.Top)
		}
		if s.Scaffold.Bottom != nil {
			newChildSpecs = append(newChildSpecs, *s.Scaffold.Bottom)
		}
		if s.Scaffold.Leading != nil {
			newChildSpecs = append(newChildSpecs, *s.Scaffold.Leading)
		}
		if s.Scaffold.Trailing != nil {
			newChildSpecs = append(newChildSpecs, *s.Scaffold.Trailing)
		}
	}
	newChildSpecs = append(newChildSpecs, s.Children...)

	for i := 0; i < len(newChildSpecs) && i < len(m.Children); i++ {
		if m.Children[i].Spec.Kind == newChildSpecs[i].Kind {
			b.Reconcile(m.Children[i], newChildSpecs[i])
		}
	}
}

func (b *FyneBridge) reconcileTabView(m *MountedNode, s ViewSpec) {
	for i := 0; i < len(s.Children) && i < len(m.Children); i++ {
		if m.Children[i].Spec.Kind == s.Children[i].Kind {
			b.Reconcile(m.Children[i], s.Children[i])
		}
	}
}

func (b *FyneBridge) reconcileNavStack(m *MountedNode, s ViewSpec) {
	// Top-of-stack view may change identity on navigation; remount if so.
	if len(s.Children) > 0 && len(m.Children) > 0 {
		if m.Children[0].Spec.Kind == s.Children[0].Kind {
			b.Reconcile(m.Children[0], s.Children[0])
		} else {
			m.Children[0].Cleanup()
			newTop := b.Mount(s.Children[0])
			m.Children[0] = newTop
			b.rebuildNavStackLayout(m, newTop.Object, s.NavBar)
			return
		}
	}
	// Rebuild nav bar so title/toolbar/back button reflect current view.
	for _, nb := range m.NavBar {
		nb.Cleanup()
	}
	var topObj fyne.CanvasObject
	if len(m.Children) > 0 {
		topObj = m.Children[0].Object
	}
	b.rebuildNavStackLayout(m, topObj, s.NavBar)
}

func (b *FyneBridge) rebuildNavStackLayout(m *MountedNode, topObj fyne.CanvasObject, bar *NavBarSpec) {
	barObj, barChildren := b.buildNavBar(bar)
	m.NavBar = barChildren
	inner := container.NewBorder(barObj, nil, nil, nil, topObj)
	if wrapper, ok := m.Object.(*fyne.Container); ok && len(wrapper.Objects) > 0 {
		wrapper.Objects[0] = inner
		wrapper.Refresh()
		return
	}
	m.Object = inner
}
