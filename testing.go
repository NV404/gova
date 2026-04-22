package gova

import "context"

// TestRender renders a component headlessly and returns a RenderedTree
// for querying and simulating interactions. No Fyne window is created.
func TestRender(root View, opts ...TestOption) *RenderedTree {
	scope := newScope(context.Background(), nil)

	for _, opt := range opts {
		opt(scope)
	}

	var node *viewNode
	if comp, ok := root.(*componentNode); ok {
		rendered := comp.renderFn(scope)
		node = rendered.viewNode()
		// Resolve nested components
		node = resolveTree(node, scope)
	} else {
		node = root.viewNode()
	}

	rt := &RenderedTree{scope: scope, root: node, component: nil}
	if comp, ok := root.(*componentNode); ok {
		rt.component = comp
	}
	return rt
}

func resolveTree(node *viewNode, scope *Scope) *viewNode {
	if node.componentRef != nil {
		rendered := node.componentRef.renderFn(scope)
		return resolveTree(rendered.viewNode(), scope)
	}
	for i, child := range node.children {
		node.children[i] = resolveTree(child, scope)
	}
	if node.scaffold != nil {
		if node.scaffold.top != nil {
			node.scaffold.top = resolveTree(node.scaffold.top.viewNode(), scope)
		}
		if node.scaffold.bottom != nil {
			node.scaffold.bottom = resolveTree(node.scaffold.bottom.viewNode(), scope)
		}
		if node.scaffold.leading != nil {
			node.scaffold.leading = resolveTree(node.scaffold.leading.viewNode(), scope)
		}
		if node.scaffold.trailing != nil {
			node.scaffold.trailing = resolveTree(node.scaffold.trailing.viewNode(), scope)
		}
	}
	return node
}

type TestOption func(*Scope)

// WithStore pre-provides a store value for isolated testing.
func WithStore[T any](key *StoreKey[T], value T) TestOption {
	return func(s *Scope) {
		Provide(s, key, value)
	}
}

// RenderedTree is the result of TestRender. It provides query methods
// to find nodes and simulate interactions, then re-render to see changes.
type RenderedTree struct {
	scope     *Scope
	root      *viewNode
	component *componentNode
}

// Rerender triggers a re-render of the component and updates the tree.
func (rt *RenderedTree) Rerender() {
	if rt.component == nil {
		return
	}
	rendered := rt.component.renderFn(rt.scope)
	rt.root = resolveTree(rendered.viewNode(), rt.scope)
}

// FindText returns the nth Text node's value (depth-first).
func (rt *RenderedTree) FindText(n int) TextResult {
	count := 0
	var result TextResult
	walkNodes(rt.root, func(node *viewNode) bool {
		if node.kind == viewKindText && node.text != nil {
			if count == n {
				switch tc := node.text.content.(type) {
				case staticText:
					result.Value = string(tc)
				case reactiveText:
					result.Value = tc.signal.Value()
				}
				result.found = true
				return false
			}
			count++
		}
		return true
	})
	return result
}

// FindButton returns the nth Button node (depth-first).
func (rt *RenderedTree) FindButton(n int) ButtonResult {
	count := 0
	var result ButtonResult
	walkNodes(rt.root, func(node *viewNode) bool {
		if node.kind == viewKindButton && node.button != nil {
			if count == n {
				result.Label = node.button.label
				result.handler = node.button.handler
				result.tree = rt
				result.found = true
				return false
			}
			count++
		}
		return true
	})
	return result
}

// FindButtonByLabel returns the first Button with a matching label.
func (rt *RenderedTree) FindButtonByLabel(label string) ButtonResult {
	var result ButtonResult
	walkNodes(rt.root, func(node *viewNode) bool {
		if node.kind == viewKindButton && node.button != nil && node.button.label == label {
			result.Label = node.button.label
			result.handler = node.button.handler
			result.tree = rt
			result.found = true
			return false
		}
		return true
	})
	return result
}

// FindToggle returns the nth Toggle node (depth-first).
func (rt *RenderedTree) FindToggle(n int) ToggleResult {
	count := 0
	var result ToggleResult
	walkNodes(rt.root, func(node *viewNode) bool {
		if node.kind == viewKindToggle && node.toggle != nil {
			if count == n {
				result.Label = node.toggle.label
				result.Checked = node.toggle.checked
				result.onChange = node.toggle.onChange
				result.tree = rt
				result.found = true
				return false
			}
			count++
		}
		return true
	})
	return result
}

// FindTextField returns the nth TextField node (depth-first).
func (rt *RenderedTree) FindTextField(n int) TextFieldResult {
	count := 0
	var result TextFieldResult
	walkNodes(rt.root, func(node *viewNode) bool {
		if node.kind == viewKindTextField && node.textField != nil {
			if count == n {
				if st, ok := node.textField.state.(*StateValue[string]); ok {
					result.state = st
				}
				result.Placeholder = node.textField.placeholder
				result.tree = rt
				result.found = true
				return false
			}
			count++
		}
		return true
	})
	return result
}

// NodeCount returns the total number of nodes of the given kind.
func (rt *RenderedTree) NodeCount(kind string) int {
	k := kindFromString(kind)
	count := 0
	walkNodes(rt.root, func(node *viewNode) bool {
		if node.kind == k {
			count++
		}
		return true
	})
	return count
}

// State retrieves a state value from the test scope by its caller key.
func (rt *RenderedTree) State(key string) any {
	rt.scope.mu.Lock()
	defer rt.scope.mu.Unlock()
	return rt.scope.states[key]
}

type TextResult struct {
	Value string
	found bool
}

func (r TextResult) Exists() bool { return r.found }

type ButtonResult struct {
	Label   string
	handler func()
	tree    *RenderedTree
	found   bool
}

func (r ButtonResult) Exists() bool { return r.found }

func (r ButtonResult) Click() {
	if r.handler != nil {
		r.handler()
	}
	if r.tree != nil {
		r.tree.Rerender()
	}
}

type ToggleResult struct {
	Label    string
	Checked  bool
	onChange func(bool)
	tree     *RenderedTree
	found    bool
}

func (r ToggleResult) Exists() bool { return r.found }

func (r ToggleResult) Toggle() {
	if r.onChange != nil {
		r.onChange(!r.Checked)
	}
	if r.tree != nil {
		r.tree.Rerender()
	}
}

type TextFieldResult struct {
	Placeholder string
	state       *StateValue[string]
	tree        *RenderedTree
	found       bool
}

func (r TextFieldResult) Exists() bool { return r.found }

func (r TextFieldResult) Text() string {
	if r.state != nil {
		return r.state.Get()
	}
	return ""
}

func (r TextFieldResult) Type(text string) {
	if r.state != nil {
		r.state.Set(text)
	}
	if r.tree != nil {
		r.tree.Rerender()
	}
}

func (r TextFieldResult) Submit() {
	if r.tree != nil {
		r.tree.Rerender()
	}
}

// walkNodes does depth-first traversal. Callback returns false to stop.
func walkNodes(node *viewNode, fn func(*viewNode) bool) bool {
	if !fn(node) {
		return false
	}
	for _, child := range node.children {
		if !walkNodes(child, fn) {
			return false
		}
	}
	if node.scaffold != nil {
		for _, edge := range []View{node.scaffold.top, node.scaffold.bottom, node.scaffold.leading, node.scaffold.trailing} {
			if edge != nil {
				if !walkNodes(edge.viewNode(), fn) {
					return false
				}
			}
		}
	}
	if node.form != nil {
		for _, item := range node.form.items {
			if !walkNodes(item.view, fn) {
				return false
			}
		}
	}
	return true
}

func kindFromString(s string) viewKind {
	switch s {
	case "Text":
		return viewKindText
	case "Button":
		return viewKindButton
	case "VStack":
		return viewKindVStack
	case "HStack":
		return viewKindHStack
	case "TextField":
		return viewKindTextField
	case "Toggle":
		return viewKindToggle
	case "List":
		return viewKindList
	case "Spacer":
		return viewKindSpacer
	case "Divider":
		return viewKindDivider
	default:
		return -1
	}
}
