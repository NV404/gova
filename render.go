package gova

import (
	"fmt"

	fyneBridge "github.com/nv404/gova/internal/fyne"
)

// themeStoreKey carries the active *Theme down the scope tree so semantic
// colors can resolve. Nil means "use built-in defaults".
var themeStoreKey = &StoreKey[*Theme]{Default: nil}

func toSpec(node *viewNode) fyneBridge.ViewSpec {
	return toSpecWithScope(node, nil)
}

func toSpecWithScope(node *viewNode, scope *Scope) fyneBridge.ViewSpec {
	if node.componentRef != nil && scope != nil {
		// A component appearing at the top of a subtree (e.g. a
		// Viewable returned directly from another component's Body)
		// gets its own child scope so its State(...) keys do not
		// collide with the caller's.
		childScope := scope.childScopeFor("body")
		rendered := renderComponent(node.componentRef, childScope)
		return toSpecWithScope(rendered.viewNode(), childScope)
	}

	var theme *Theme
	if scope != nil {
		theme = UseStore(scope, themeStoreKey).Get()
	}

	spec := fyneBridge.ViewSpec{Key: node.key}

	switch node.kind {
	case viewKindText:
		spec.Kind = fyneBridge.KindText
		strike := node.modifier.strikethrough
		if node.text != nil {
			switch tc := node.text.content.(type) {
			case staticText:
				spec.TextValue = applyStrikethrough(string(tc), strike)
			case reactiveText:
				spec.TextValue = applyStrikethrough(tc.signal.Value(), strike)
				spec.TextSubscribe = func(fn func(string)) func() {
					return tc.signal.subscribe(func(s string) { fn(applyStrikethrough(s, strike)) })
				}
			}
		}
		spec.Bold = node.modifier.bold
		spec.Italic = node.modifier.italic
		if node.modifier.font != nil {
			spec.FontStyle = int(*node.modifier.font)
		}
		if c := resolveColor(node.modifier.color, theme); c != nil {
			r, g, b, a := c.RGBA()
			spec.TextColor = [4]uint8{uint8(r >> 8), uint8(g >> 8), uint8(b >> 8), uint8(a >> 8)}
			spec.HasColor = true
		}

	case viewKindButton:
		spec.Kind = fyneBridge.KindButton
		if node.button != nil {
			spec.ButtonLabel = node.button.label
			spec.ButtonHandler = node.button.handler
		}

	case viewKindVStack:
		spec.Kind = fyneBridge.KindVStack
		spec.Children = childSpecsWithScope(node, scope)
		applyLayout(&spec, node)

	case viewKindHStack:
		spec.Kind = fyneBridge.KindHStack
		spec.Children = childSpecsWithScope(node, scope)
		applyLayout(&spec, node)

	case viewKindZStack:
		spec.Kind = fyneBridge.KindZStack
		spec.Children = childSpecsWithScope(node, scope)
		applyLayout(&spec, node)

	case viewKindGroup:
		spec.Kind = fyneBridge.KindGroup
		spec.Children = childSpecsWithScope(node, scope)

	case viewKindSpacer:
		spec.Kind = fyneBridge.KindSpacer

	case viewKindTextField:
		spec.Kind = fyneBridge.KindTextField
		if node.textField != nil {
			if state, ok := node.textField.state.(*StateValue[string]); ok {
				spec.TextFieldGet = state.Get
				spec.TextFieldSet = state.Set
				spec.TextFieldSetSilent = state.SetSilent
			}
			spec.TextFieldPlaceholder = node.textField.placeholder
			spec.TextFieldOnSubmit = node.textField.onSubmit
			spec.TextFieldMultiline = node.textField.multiline
			spec.TextFieldPassword = node.textField.password
		}

	case viewKindToggle:
		spec.Kind = fyneBridge.KindToggle
		if node.toggle != nil {
			spec.ToggleLabel = node.toggle.label
			spec.ToggleChecked = node.toggle.checked
			if st, ok := node.toggle.state.(*StateValue[bool]); ok {
				spec.ToggleChecked = st.Get()
				userFn := node.toggle.onChange
				spec.ToggleOnChange = func(v bool) {
					st.Set(v)
					if userFn != nil {
						userFn(v)
					}
				}
			} else {
				spec.ToggleOnChange = node.toggle.onChange
			}
		}

	case viewKindSlider:
		spec.Kind = fyneBridge.KindSlider
		if node.slider != nil {
			spec.SliderMin = node.slider.min
			spec.SliderMax = node.slider.max
			spec.SliderValue = node.slider.value
			spec.SliderStep = node.slider.step
			if !node.slider.hasRange {
				spec.SliderMin = 0
				spec.SliderMax = 1
			}
			if st, ok := node.slider.state.(*StateValue[float64]); ok {
				spec.SliderValue = st.Get()
				userFn := node.slider.onChange
				spec.SliderOnChange = func(v float64) {
					st.Set(v)
					if userFn != nil {
						userFn(v)
					}
				}
			} else {
				spec.SliderOnChange = node.slider.onChange
			}
		}

	case viewKindPicker:
		spec.Kind = fyneBridge.KindPicker
		if node.picker != nil {
			spec.PickerOptions = node.picker.options
			spec.PickerSelected = node.picker.selected
			spec.PickerOnChange = node.picker.onChange
		}

	case viewKindProgress:
		spec.Kind = fyneBridge.KindProgress
		if node.progress != nil {
			spec.ProgressValue = node.progress.value
		}

	case viewKindList:
		spec.Kind = fyneBridge.KindList
		spec.Children = childSpecsWithScope(node, scope)

	case viewKindGrid:
		spec.Kind = fyneBridge.KindGrid
		spec.Children = childSpecsWithScope(node, scope)
		if node.grid != nil {
			spec.GridColumns = node.grid.columns
		}

	case viewKindForm:
		spec.Kind = fyneBridge.KindForm
		if node.form != nil {
			spec.FormItems = make([]fyneBridge.FormItemSpec, len(node.form.items))
			for i, item := range node.form.items {
				spec.FormItems[i] = fyneBridge.FormItemSpec{
					Label: item.label,
					View:  renderSlot(item.view, scope, fmt.Sprintf("form:%d", i)),
				}
			}
		}

	case viewKindScrollView:
		spec.Kind = fyneBridge.KindScrollView
		spec.Children = childSpecsWithScope(node, scope)
		if node.scrollView != nil {
			spec.ScrollDirection = int(node.scrollView.direction)
		}

	case viewKindDivider:
		spec.Kind = fyneBridge.KindDivider

	case viewKindScaffold:
		spec.Kind = fyneBridge.KindScaffold
		spec.Children = childSpecsWithScope(node, scope)
		if node.scaffold != nil {
			spec.Scaffold = &fyneBridge.ScaffoldSpec{}
			if node.scaffold.top != nil {
				s := renderSlot(node.scaffold.top.viewNode(), scope, "scaffold:top")
				spec.Scaffold.Top = &s
			}
			if node.scaffold.bottom != nil {
				s := renderSlot(node.scaffold.bottom.viewNode(), scope, "scaffold:bottom")
				spec.Scaffold.Bottom = &s
			}
			if node.scaffold.leading != nil {
				s := renderSlot(node.scaffold.leading.viewNode(), scope, "scaffold:leading")
				spec.Scaffold.Leading = &s
			}
			if node.scaffold.trailing != nil {
				s := renderSlot(node.scaffold.trailing.viewNode(), scope, "scaffold:trailing")
				spec.Scaffold.Trailing = &s
			}
		}

	case viewKindTabView:
		spec.Kind = fyneBridge.KindTabView
		spec.Children = childSpecsWithScope(node, scope)
		if node.tabView != nil {
			spec.TabLabels = make([]string, len(node.tabView.items))
			spec.TabIcons = make([]string, len(node.tabView.items))
			spec.TabPlacement = int(node.tabView.placement)
			for i, item := range node.tabView.items {
				spec.TabLabels[i] = item.label
				spec.TabIcons[i] = item.icon
			}
		}

	case viewKindImage:
		spec.Kind = fyneBridge.KindImage
		if node.image != nil {
			spec.ImageFilePath = node.image.filePath
			spec.ImageURL = node.image.url
			spec.ImageFillMode = int(node.image.fillMode)
		}

	case viewKindNavStack:
		return navStackSpec(node, scope, theme)

	case viewKindComponent:
		spec.Kind = fyneBridge.KindGroup
	}

	if node.modifier.padding != nil {
		spec.HasPadding = true
		spec.PaddingTop = node.modifier.padding.top
		spec.PaddingRight = node.modifier.padding.right
		spec.PaddingBottom = node.modifier.padding.bottom
		spec.PaddingLeft = node.modifier.padding.left
	}

	if node.modifier.onTap != nil {
		spec.OnTap = node.modifier.onTap
		spec.HasOnTap = true
	}
	if node.modifier.hasCorner {
		spec.CornerRadius = node.modifier.cornerRadius
		spec.HasCornerRadius = true
	}
	if node.modifier.shadow != nil {
		if c := resolveColor(node.modifier.shadow.color, theme); c != nil {
			r, g, b, a := c.RGBA()
			spec.ShadowColor = [4]uint8{uint8(r >> 8), uint8(g >> 8), uint8(b >> 8), uint8(a >> 8)}
		}
		spec.ShadowRadius = node.modifier.shadow.radius
		spec.ShadowOffsetX = node.modifier.shadow.offsetX
		spec.ShadowOffsetY = node.modifier.shadow.offsetY
		spec.HasShadow = true
	}
	if node.modifier.stroke != nil {
		if c := resolveColor(node.modifier.stroke.color, theme); c != nil {
			r, g, b, a := c.RGBA()
			spec.StrokeColor = [4]uint8{uint8(r >> 8), uint8(g >> 8), uint8(b >> 8), uint8(a >> 8)}
		}
		spec.StrokeWidth = node.modifier.stroke.width
		spec.HasStroke = true
	}
	if node.modifier.background != nil {
		if c := resolveColor(node.modifier.background, theme); c != nil {
			r, g, b, a := c.RGBA()
			spec.BackgroundColor = [4]uint8{uint8(r >> 8), uint8(g >> 8), uint8(b >> 8), uint8(a >> 8)}
			spec.HasBackground = true
		}
	}
	if node.modifier.hasOpacity {
		spec.Opacity = node.modifier.opacity
		spec.HasOpacity = true
	}
	if node.modifier.hasA11y {
		spec.A11yLabel = node.modifier.a11yLabel
		spec.A11yHint = node.modifier.a11yHint
		spec.A11yRole = int(node.modifier.a11yRole)
		spec.HasA11y = true
	}
	if node.modifier.frame != nil {
		fm := node.modifier.frame
		if fm.width > 0 || fm.height > 0 {
			spec.MinWidth = fm.width
			spec.MinHeight = fm.height
			spec.HasMinSize = true
		}
		if fm.minWidth > 0 && fm.minWidth > spec.MinWidth {
			spec.MinWidth = fm.minWidth
			spec.HasMinSize = true
		}
		if fm.minHeight > 0 && fm.minHeight > spec.MinHeight {
			spec.MinHeight = fm.minHeight
			spec.HasMinSize = true
		}
	}
	spec.Grow = node.modifier.grow

	return spec
}

// applyStrikethrough adds Unicode combining long stroke overlay (\u0336)
// after every rune when enabled. This renders as a strikethrough in any
// font without needing renderer-specific support.
func applyStrikethrough(s string, on bool) string {
	if !on || s == "" {
		return s
	}
	out := make([]rune, 0, len(s)*2)
	for _, r := range s {
		out = append(out, r, '\u0336')
	}
	return string(out)
}

func applyLayout(spec *fyneBridge.ViewSpec, node *viewNode) {
	if node.layout == nil {
		return
	}
	spec.Spacing = node.layout.spacing
	spec.HasSpacing = node.layout.spacing > 0
	if node.layout.hasAlignment {
		spec.Alignment = int(node.layout.alignment)
		spec.HasAlignment = true
	}
}

func childSpecsWithScope(node *viewNode, scope *Scope) []fyneBridge.ViewSpec {
	specs := make([]fyneBridge.ViewSpec, len(node.children))
	for i, child := range node.children {
		specs[i] = renderSlot(child, scope, childSlotID(child, i))
	}
	return specs
}

// renderSlot is the single entry point for rendering a node that *might*
// be a component reference sitting in a named slot (a child position, a
// scaffold slot, a form row, a nav-stack top, etc.). If the node wraps a
// component, the component is rendered inside a per-slot child scope so
// sibling components of the same type hold independent state.
func renderSlot(node *viewNode, parentScope *Scope, slotID string) fyneBridge.ViewSpec {
	if node.componentRef != nil && parentScope != nil {
		childScope := parentScope.childScopeFor(slotID)
		rendered := renderComponent(node.componentRef, childScope)
		return toSpecWithScope(rendered.viewNode(), childScope)
	}
	return toSpecWithScope(node, parentScope)
}

// childSlotID produces a stable ID for a child node within its parent. An
// explicit Key() wins; otherwise state follows the child's position. This
// mirrors React's default behavior for keyed vs positional identity.
func childSlotID(node *viewNode, index int) string {
	if node.key != nil {
		return fmt.Sprintf("child:key:%v", node.key)
	}
	return fmt.Sprintf("child:pos:%d", index)
}

// navStackSpec renders a NavStack: provides the Nav handle into the scope,
// then renders only the top-of-stack view with the nav bar wrapping it.
//
// The stack state must persist across renders. The viewNode is recreated on
// every re-render (because `g.NavStack(...)` allocates fresh), so we stash
// the Nav in the scope's store on first mount and reuse it thereafter.
func navStackSpec(node *viewNode, scope *Scope, theme *Theme) fyneBridge.ViewSpec {
	if node.nav == nil || scope == nil {
		if len(node.children) > 0 {
			return renderSlot(node.children[0], scope, "nav:root")
		}
		return fyneBridge.ViewSpec{Kind: fyneBridge.KindGroup}
	}

	nav := UseStore(scope, navStackStoreKey).Get()
	if nav == nil {
		initial := make([]View, 0, 1)
		if len(node.children) > 0 {
			initial = append(initial, node.children[0])
		}
		nav = &Nav{stack: newState(initial, scope.onStateChange)}
		// Write the Nav directly into the scope's store slot. Provide() is
		// idempotent once the slot exists (UseStore above creates it), so
		// we must set the value imperatively. SetSilent avoids triggering
		// a re-render from inside a render.
		UseStore(scope, navStackStoreKey).SetSilent(nav)
	}
	node.nav.stack = nav.stack

	stack := nav.stack.Get()
	var top *viewNode
	if len(stack) > 0 {
		top = stack[len(stack)-1].viewNode()
	} else if len(node.children) > 0 {
		top = node.children[0]
	} else {
		return fyneBridge.ViewSpec{Kind: fyneBridge.KindGroup}
	}

	// Resolve the top view (it may be a component) to capture its nav modifiers.
	topSpec := renderSlot(top, scope, "nav:top")
	resolvedTop := resolveNavModifiers(top, scope)

	barTitle := ""
	var toolbar *navToolbarData
	if resolvedTop != nil {
		if resolvedTop.modifier.hasNavTitle {
			barTitle = resolvedTop.modifier.navTitle
		}
		if resolvedTop.modifier.navToolbar != nil {
			toolbar = resolvedTop.modifier.navToolbar
		}
	}

	bar := fyneBridge.NavBarSpec{
		Title:  barTitle,
		CanPop: nav.CanPop(),
		OnBack: nav.Pop,
	}
	if toolbar != nil {
		for i, v := range toolbar.leading {
			s := renderSlot(v.viewNode(), scope, fmt.Sprintf("toolbar:leading:%d", i))
			bar.Leading = append(bar.Leading, s)
		}
		for i, v := range toolbar.trailing {
			s := renderSlot(v.viewNode(), scope, fmt.Sprintf("toolbar:trailing:%d", i))
			bar.Trailing = append(bar.Trailing, s)
		}
	}

	return fyneBridge.ViewSpec{
		Kind:     fyneBridge.KindNavStack,
		Children: []fyneBridge.ViewSpec{topSpec},
		NavBar:   &bar,
	}
}

// resolveNavModifiers walks a (possibly component) node to find the final
// viewNode whose modifiers carry NavTitle / NavToolbar.
func resolveNavModifiers(node *viewNode, scope *Scope) *viewNode {
	cur := node
	for cur != nil && cur.componentRef != nil && scope != nil {
		rendered := renderComponent(cur.componentRef, scope)
		cur = rendered.viewNode()
	}
	return cur
}
