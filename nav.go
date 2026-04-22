package gova

import "sync"

// TabPlacement controls where a TabView renders its tab bar.
type TabPlacement int

const (
	TabTop TabPlacement = iota
	TabBottom
	TabLeading
	TabTrailing
)

type tabItem struct {
	label   string
	icon    string
	content View
}

type tabViewData struct {
	items     []tabItem
	placement TabPlacement
}

// TabView creates a tabbed container. Each Tab is a label + optional icon
// path + content.
func TabView(tabs ...TabEntry) *viewNode {
	items := make([]tabItem, len(tabs))
	children := make([]*viewNode, len(tabs))
	for i, t := range tabs {
		items[i] = tabItem{label: t.Label, icon: t.Icon, content: t.Content}
		children[i] = t.Content.viewNode()
	}
	return &viewNode{
		kind:     viewKindTabView,
		children: children,
		tabView:  &tabViewData{items: items},
	}
}

type TabEntry struct {
	Label   string
	Icon    string
	Content View
}

func Tab(label, icon string, content View) TabEntry {
	return TabEntry{Label: label, Icon: icon, Content: content}
}

func (n *viewNode) Placement(p TabPlacement) *viewNode {
	if n.tabView != nil {
		n.tabView.placement = p
	}
	return n
}

// --- Stack navigation ---

// navData holds the state of a NavStack. Stored on the viewNode so the
// reconciler can look it up; also referenced by the overlay store so
// UseNav can find it from descendant scopes.
type navData struct {
	stack *StateValue[[]View]
}

// navStackStoreKey is an internal store used to pass the active Nav handle
// down the scope tree to UseNav callers. Allocated once.
var navStackStoreKey = &StoreKey[*Nav]{Default: nil}

// Nav is the runtime handle to a navigation stack. Obtain one via UseNav
// inside a view rendered under a NavStack.
type Nav struct {
	mu    sync.Mutex
	stack *StateValue[[]View]
}

// Push appends a view to the stack; it becomes the visible screen.
// Accepts a View or a Viewable (pass a struct literal like DetailView{ID: 42}).
func (n *Nav) Push(v any) {
	if n == nil {
		return
	}
	view := asView(v)
	if view == nil {
		return
	}
	n.mu.Lock()
	defer n.mu.Unlock()
	n.stack.Update(func(s []View) []View {
		return append(s, view)
	})
}

// Pop removes the top of the stack. No-op if the root is the only entry.
func (n *Nav) Pop() {
	if n == nil {
		return
	}
	n.mu.Lock()
	defer n.mu.Unlock()
	n.stack.Update(func(s []View) []View {
		if len(s) <= 1 {
			return s
		}
		return s[:len(s)-1]
	})
}

// PopToRoot removes all but the root view.
func (n *Nav) PopToRoot() {
	if n == nil {
		return
	}
	n.mu.Lock()
	defer n.mu.Unlock()
	n.stack.Update(func(s []View) []View {
		if len(s) == 0 {
			return s
		}
		return s[:1]
	})
}

// Replace replaces the top of the stack in place (no push).
func (n *Nav) Replace(v any) {
	if n == nil {
		return
	}
	view := asView(v)
	if view == nil {
		return
	}
	n.mu.Lock()
	defer n.mu.Unlock()
	n.stack.Update(func(s []View) []View {
		if len(s) == 0 {
			return []View{view}
		}
		out := append([]View(nil), s...)
		out[len(out)-1] = view
		return out
	})
}

// CanPop returns true if there is more than just the root view on the stack.
func (n *Nav) CanPop() bool {
	if n == nil {
		return false
	}
	return len(n.stack.Get()) > 1
}

// Depth returns the number of views on the stack.
func (n *Nav) Depth() int {
	if n == nil {
		return 0
	}
	return len(n.stack.Get())
}

// NavStack creates a navigation container rooted at the given view.
// Inside any view in the stack, call UseNav(s) to push/pop.
//
//	g.NavStack(HomeView{})
func NavStack(root any) *viewNode {
	rootView := asView(root)
	if rootView == nil {
		panic("gova: NavStack requires a non-nil root view")
	}
	return &viewNode{
		kind: viewKindNavStack,
		children: []*viewNode{rootView.viewNode()},
		nav:  &navData{},
	}
}

// UseNav returns the Nav handle for the enclosing NavStack.
// Returns a no-op Nav if called outside a NavStack.
func UseNav(s *Scope) *Nav {
	store := UseStore(s, navStackStoreKey)
	nav := store.Get()
	if nav == nil {
		return &Nav{stack: &StateValue[[]View]{}}
	}
	return nav
}

// NavLink renders as a tappable label; on tap it pushes dest onto the
// enclosing NavStack. dest may be a View, a Viewable, or a func() View
// for fully lazy construction.
func NavLink(label string, dest any) View {
	return Define(func(s *Scope) View {
		nav := UseNav(s)
		handler := func() {
			switch d := dest.(type) {
			case func() View:
				nav.Push(d())
			case func() any:
				nav.Push(d())
			default:
				nav.Push(d)
			}
		}
		return Text(label).
			Color(Accent).
			OnTap(handler).
			AccessibilityRole(RoleLink).
			AccessibilityLabel(label)
	})
}

// NavTitle sets the title shown by the NavStack's title bar for the
// currently visible view. Apply on the view returned by Body.
func (n *viewNode) NavTitle(title string) *viewNode {
	n.modifier.navTitle = title
	n.modifier.hasNavTitle = true
	return n
}

// NavToolbar attaches toolbar items to the NavStack's bar for the
// currently visible view. Items are ToolbarItem values produced by
// ToolbarLeading / ToolbarTrailing.
func (n *viewNode) NavToolbar(items ...ToolbarItem) *viewNode {
	tb := &navToolbarData{}
	for _, it := range items {
		switch it.placement {
		case toolbarLeading:
			tb.leading = append(tb.leading, it.view)
		case toolbarTrailing:
			tb.trailing = append(tb.trailing, it.view)
		}
	}
	n.modifier.navToolbar = tb
	return n
}

type navToolbarData struct {
	leading  []View
	trailing []View
}

type toolbarPlacement int

const (
	toolbarLeading toolbarPlacement = iota
	toolbarTrailing
)

// ToolbarItem is an entry in NavToolbar. Build with ToolbarLeading or
// ToolbarTrailing.
type ToolbarItem struct {
	placement toolbarPlacement
	view      View
}

func ToolbarLeading(v any) ToolbarItem {
	return ToolbarItem{placement: toolbarLeading, view: asView(v)}
}

func ToolbarTrailing(v any) ToolbarItem {
	return ToolbarItem{placement: toolbarTrailing, view: asView(v)}
}
