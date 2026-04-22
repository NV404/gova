package gova

import (
	"context"
	"fmt"
	"testing"
)

func TestStateGetSet(t *testing.T) {
	scope := newScope(context.Background(), nil)

	// Simulate what g.State(s, 0) does - manually specify key since
	// runtime.Caller won't give stable keys in tests
	st := getOrCreateState(scope, "test:count", 0)

	if st.Get() != 0 {
		t.Fatalf("expected 0, got %d", st.Get())
	}

	st.Set(5)
	if st.Get() != 5 {
		t.Fatalf("expected 5, got %d", st.Get())
	}

	st.Update(func(n int) int { return n + 10 })
	if st.Get() != 15 {
		t.Fatalf("expected 15, got %d", st.Get())
	}
}

func TestStateNotifiesOnChange(t *testing.T) {
	called := false
	scope := newScope(context.Background(), func() { called = true })

	st := getOrCreateState(scope, "test:x", 0)
	st.Set(1)

	if !called {
		t.Fatal("expected onStateChange callback to fire")
	}
}

func TestRefDoesNotNotify(t *testing.T) {
	called := false
	scope := newScope(context.Background(), func() { called = true })

	r := getOrCreateRef(scope, "test:r", 0)
	r.Set(42)

	if called {
		t.Fatal("Ref.Set should not trigger onStateChange")
	}
	if r.Get() != 42 {
		t.Fatalf("expected 42, got %d", r.Get())
	}
}

func TestFormatSignal(t *testing.T) {
	scope := newScope(context.Background(), nil)
	st := getOrCreateState(scope, "test:n", 5)

	sig := st.Format("Value: %d")
	if sig.Value() != "Value: 5" {
		t.Fatalf("expected 'Value: 5', got %q", sig.Value())
	}

	st.Set(10)
	if sig.Value() != "Value: 10" {
		t.Fatalf("expected 'Value: 10', got %q", sig.Value())
	}
}

func TestFormatSignalSubscribe(t *testing.T) {
	scope := newScope(context.Background(), nil)
	st := getOrCreateState(scope, "test:n", 0)
	sig := st.Format("Count: %d")

	var received string
	sig.subscribe(func(s string) { received = s })

	st.Set(7)
	if received != "Count: 7" {
		t.Fatalf("expected 'Count: 7', got %q", received)
	}
}

func TestViewTreeConstruction(t *testing.T) {
	tree := VStack(
		Text("hello"),
		Button("+", func() {}),
		HStack(
			Text("nested"),
			Spacer(),
		),
	)

	node := tree.viewNode()
	if node.kind != viewKindVStack {
		t.Fatal("expected VStack")
	}
	if len(node.children) != 3 {
		t.Fatalf("expected 3 children, got %d", len(node.children))
	}
	if node.children[0].kind != viewKindText {
		t.Fatal("expected first child to be Text")
	}
	if node.children[1].kind != viewKindButton {
		t.Fatal("expected second child to be Button")
	}
	if node.children[2].kind != viewKindHStack {
		t.Fatal("expected third child to be HStack")
	}
	if len(node.children[2].children) != 2 {
		t.Fatalf("expected HStack to have 2 children, got %d", len(node.children[2].children))
	}
}

func TestNilChildrenSkipped(t *testing.T) {
	tree := VStack(nil, Text("ok"), nil)
	node := tree.viewNode()
	if len(node.children) != 1 {
		t.Fatalf("expected 1 non-nil child, got %d", len(node.children))
	}
}

func TestDerivedSignal(t *testing.T) {
	scope := newScope(context.Background(), nil)
	st := getOrCreateState(scope, "test:n", 3)

	doubled := Derived(st, func(n int) int { return n * 2 })
	if doubled.Value() != 6 {
		t.Fatalf("expected 6, got %d", doubled.Value())
	}

	st.Set(10)
	if doubled.Value() != 20 {
		t.Fatalf("expected 20, got %d", doubled.Value())
	}
}

func TestStateReusesAcrossRenders(t *testing.T) {
	scope := newScope(context.Background(), nil)

	st1 := getOrCreateState(scope, "test:count", 0)
	st1.Set(42)

	st2 := getOrCreateState(scope, "test:count", 0)
	if st2.Get() != 42 {
		t.Fatalf("expected existing state with value 42, got %d", st2.Get())
	}

	if st1 != st2 {
		t.Fatal("expected same State instance across renders")
	}
}

func TestTextFieldViewNode(t *testing.T) {
	scope := newScope(context.Background(), nil)
	st := getOrCreateState(scope, "test:input", "hello")

	node := TextField(st).Placeholder("type...").viewNode()
	if node.kind != viewKindTextField {
		t.Fatal("expected TextField kind")
	}
	if node.textField.placeholder != "type..." {
		t.Fatalf("expected placeholder 'type...', got %q", node.textField.placeholder)
	}
}

func TestToggleViewNode(t *testing.T) {
	var toggled bool
	node := Toggle(true).Label("Enable").OnChange(func(v bool) { toggled = v }).viewNode()
	if node.kind != viewKindToggle {
		t.Fatal("expected Toggle kind")
	}
	if !node.toggle.checked {
		t.Fatal("expected checked=true")
	}
	node.toggle.onChange(false)
	if toggled {
		t.Fatal("expected toggled=false after onChange(false)")
	}
}

func TestListCreatesChildren(t *testing.T) {
	scope := newScope(context.Background(), nil)
	items := getOrCreateState(scope, "test:items", []string{"a", "b", "c"})

	node := List(items, func(s string) string { return s }, func(i int, s string) View {
		return Text(s)
	}).viewNode()

	if len(node.children) != 3 {
		t.Fatalf("expected 3 children, got %d", len(node.children))
	}
}

func TestForEachFlattensIntoGroup(t *testing.T) {
	items := []int{1, 2, 3}
	node := ForEachIndexed(items, func(i int, n int) any {
		return Text("item")
	}).viewNode()

	if node.kind != viewKindGroup {
		t.Fatal("expected Group kind")
	}
	if len(node.children) != 3 {
		t.Fatalf("expected 3 children, got %d", len(node.children))
	}
}

func TestWhenTrue(t *testing.T) {
	node := When(true, func() View { return Text("yes") }).viewNode()
	if node.kind != viewKindText {
		t.Fatal("expected Text when condition is true")
	}
}

func TestWhenFalse(t *testing.T) {
	node := When(false, func() View { return Text("no") }).viewNode()
	if node.kind != viewKindGroup {
		t.Fatal("expected empty Group when condition is false")
	}
}

func TestDerivedList(t *testing.T) {
	type Model struct {
		Items  []string
		NextID int
	}
	scope := newScope(context.Background(), nil)
	model := getOrCreateState(scope, "test:model", Model{Items: []string{"a", "b"}, NextID: 3})

	derived := DerivedList(model, func(m Model) []string { return m.Items })
	if len(derived.Get()) != 2 {
		t.Fatalf("expected 2 items, got %d", len(derived.Get()))
	}

	model.Update(func(m Model) Model {
		m.Items = append(m.Items, "c")
		return m
	})
	if len(derived.Get()) != 3 {
		t.Fatalf("expected 3 items after update, got %d", len(derived.Get()))
	}
}

func TestUseEffectRunsOnce(t *testing.T) {
	count := 0
	scope := newScope(context.Background(), nil)

	runEffect := func() {
		UseEffect(scope, func(ctx context.Context) func() {
			count++
			return nil
		})
	}

	runEffect() // first call - should run
	runEffect() // same call site - should be a no-op
	if count != 1 {
		t.Fatalf("expected effect to run once, ran %d times", count)
	}
}

func TestFormViewNode(t *testing.T) {
	scope := newScope(context.Background(), nil)
	input := getOrCreateState(scope, "test:name", "")

	node := Form(
		FormEntry{Label: "Name", View: TextField(input)},
		FormEntry{Label: "Bio", View: TextField(input).Multiline()},
	).viewNode()

	if node.kind != viewKindForm {
		t.Fatal("expected Form kind")
	}
	if len(node.form.items) != 2 {
		t.Fatalf("expected 2 form items, got %d", len(node.form.items))
	}
}

func TestGridViewNode(t *testing.T) {
	node := Grid(3, Text("a"), Text("b"), Text("c")).viewNode()
	if node.kind != viewKindGrid {
		t.Fatal("expected Grid kind")
	}
	if node.grid.columns != 3 {
		t.Fatalf("expected 3 columns, got %d", node.grid.columns)
	}
	if len(node.children) != 3 {
		t.Fatalf("expected 3 children, got %d", len(node.children))
	}
}

func TestScrollViewNode(t *testing.T) {
	node := ScrollView(Text("content")).viewNode()
	if node.kind != viewKindScrollView {
		t.Fatal("expected ScrollView kind")
	}
	if len(node.children) != 1 {
		t.Fatalf("expected 1 child, got %d", len(node.children))
	}
}

func TestProvideAndUseStore(t *testing.T) {
	key := &StoreKey[string]{Default: "default"}
	scope := newScope(context.Background(), nil)

	Provide(scope, key, "provided")
	store := UseStore(scope, key)
	if store.Get() != "provided" {
		t.Fatalf("expected 'provided', got %q", store.Get())
	}

	store.Set("updated")
	if store.Get() != "updated" {
		t.Fatalf("expected 'updated', got %q", store.Get())
	}
}

func TestStoreInheritsFromParent(t *testing.T) {
	key := &StoreKey[int]{Default: 0}
	parent := newScope(context.Background(), nil)
	Provide(parent, key, 42)

	child := newChildScope(parent, nil)
	store := UseStore(child, key)
	if store.Get() != 42 {
		t.Fatalf("expected 42 from parent, got %d", store.Get())
	}
}

func TestStoreFallsBackToDefault(t *testing.T) {
	key := &StoreKey[string]{Default: "fallback"}
	scope := newScope(context.Background(), nil)

	store := UseStore(scope, key)
	if store.Get() != "fallback" {
		t.Fatalf("expected 'fallback', got %q", store.Get())
	}
}

func TestStoreIsolationBetweenSiblings(t *testing.T) {
	key := &StoreKey[int]{Default: 0}
	root := newScope(context.Background(), nil)

	child1 := newChildScope(root, nil)
	Provide(child1, key, 100)

	child2 := newChildScope(root, nil)
	store2 := UseStore(child2, key)
	// child2 shouldn't see child1's provided value
	if store2.Get() != 0 {
		t.Fatalf("expected default 0, got %d (leaked from sibling)", store2.Get())
	}
}

func TestTabViewNode(t *testing.T) {
	node := TabView(
		Tab("Home", "", Text("home content")),
		Tab("Settings", "", Text("settings content")),
	).Placement(TabBottom).viewNode()

	if node.kind != viewKindTabView {
		t.Fatal("expected TabView kind")
	}
	if len(node.children) != 2 {
		t.Fatalf("expected 2 tabs, got %d", len(node.children))
	}
	if node.tabView.placement != TabBottom {
		t.Fatal("expected TabBottom placement")
	}
}

func TestNavigationPushPop(t *testing.T) {
	scope := newScope(context.Background(), nil)
	st := getOrCreateState(scope, "nav:test", []View{Text("root")})
	nav := &Nav{stack: st}

	if len(nav.stack.Get()) != 1 {
		t.Fatal("expected 1 item in stack")
	}

	nav.Push(Text("detail"))
	if len(nav.stack.Get()) != 2 {
		t.Fatal("expected 2 items after push")
	}
	if !nav.CanPop() {
		t.Fatal("expected CanPop=true")
	}

	nav.Pop()
	if len(nav.stack.Get()) != 1 {
		t.Fatal("expected 1 item after pop")
	}

	nav.Pop()
	if len(nav.stack.Get()) != 1 {
		t.Fatal("expected 1 item after extra pop")
	}
}

func TestErrorBoundary(t *testing.T) {
	panicking := Define(func(s *Scope) View {
		panic("test panic")
	})

	_ = ErrorBoundary(panicking, func(err error) View {
		return Text("caught: " + err.Error())
	})

	comp := panicking.(*componentNode)
	scope := newScope(context.Background(), nil)
	result := comp.renderFn(scope)
	node := result.viewNode()

	if node.kind != viewKindText {
		t.Fatal("expected fallback Text view")
	}
}

func TestImageNode(t *testing.T) {
	node := Image("/path/to/img.png").Fill(FillCover).viewNode()
	if node.kind != viewKindImage {
		t.Fatal("expected Image kind")
	}
	if node.image.filePath != "/path/to/img.png" {
		t.Fatalf("expected file path, got %q", node.image.filePath)
	}
	if node.image.fillMode != FillCover {
		t.Fatal("expected FillCover")
	}
}

func TestScaffoldLayout(t *testing.T) {
	node := Scaffold(Text("center")).
		Top(Text("top")).
		Bottom(Text("bottom")).
		Trailing(Text("right")).
		viewNode()
	if node.kind != viewKindScaffold {
		t.Fatal("expected Scaffold kind")
	}
	if node.scaffold.top == nil {
		t.Fatal("expected top view")
	}
	if node.scaffold.leading != nil {
		t.Fatal("expected nil leading view")
	}
	if node.scaffold.trailing == nil {
		t.Fatal("expected trailing view")
	}
	if len(node.children) != 1 {
		t.Fatalf("expected 1 center child, got %d", len(node.children))
	}
}

func TestNestedComponentRendering(t *testing.T) {
	inner := Define(func(s *Scope) View {
		return Text("from inner component")
	})

	outer := Define(func(s *Scope) View {
		return VStack(inner, Text("static"))
	})

	scope := newScope(context.Background(), nil)
	comp := outer.(*componentNode)
	view := comp.renderFn(scope)
	spec := toSpecWithScope(view.viewNode(), scope)

	if len(spec.Children) != 2 {
		t.Fatalf("expected 2 children, got %d", len(spec.Children))
	}
	if spec.Children[0].TextValue != "from inner component" {
		t.Fatalf("expected inner component text, got %q", spec.Children[0].TextValue)
	}
}

func TestNestedComponentWithStore(t *testing.T) {
	key := &StoreKey[string]{Default: "default"}

	inner := Define(func(s *Scope) View {
		store := UseStore(s, key)
		return Text(store.Get())
	})

	scope := newScope(context.Background(), nil)
	Provide(scope, key, "shared value")

	outer := Define(func(s *Scope) View {
		return inner
	})

	comp := outer.(*componentNode)
	view := comp.renderFn(scope)
	spec := toSpecWithScope(view.viewNode(), scope)

	if spec.TextValue != "shared value" {
		t.Fatalf("expected 'shared value', got %q", spec.TextValue)
	}
}

func TestHexColor(t *testing.T) {
	c := Hex("#FF5733")
	r, g, b, a := c.RGBA()
	// NRGBA values are pre-multiplied in RGBA()
	if r>>8 != 255 || g>>8 != 87 || b>>8 != 51 || a>>8 != 255 {
		t.Fatalf("unexpected color: R=%d G=%d B=%d A=%d", r>>8, g>>8, b>>8, a>>8)
	}
}

func TestHexColorWithAlpha(t *testing.T) {
	c := Hex("#FF573380")
	_, _, _, a := c.RGBA()
	if a>>8 != 128 {
		t.Fatalf("expected alpha ~128, got %d", a>>8)
	}
}

func TestThemeBuilder(t *testing.T) {
	theme := DarkTheme().WithAccent(Purple).SetColor(ColorError, Red)

	if theme.Variant != ThemeDark {
		t.Fatal("expected dark variant")
	}
	if theme.Accent == nil {
		t.Fatal("expected accent color")
	}
	if _, ok := theme.Colors[ColorError]; !ok {
		t.Fatal("expected error color override")
	}
}

func TestUseAlertWithoutOverlays(t *testing.T) {
	scope := newScope(context.Background(), nil)
	showAlert := UseAlert(scope)
	// Should not panic when overlay store is nil (no window)
	showAlert("Title", "Message")
}

func TestUseSheetWithoutOverlays(t *testing.T) {
	scope := newScope(context.Background(), nil)
	show, dismiss := UseSheet(scope)
	// Should not panic when overlay store is nil (no window)
	show(Text("content"))
	dismiss()
}

// internalSubCount reports how many callbacks the source state is
// currently fanning out to. Tests use it to guard against subscriber
// leaks: repeated renders from the same call site must not grow this.
func internalSubCount[T any](st *StateValue[T]) int {
	st.internalSubsMu.Lock()
	defer st.internalSubsMu.Unlock()
	return len(st.internalSubs)
}

// A standalone helper keeps the Format call anchored at one file:line so
// the caller-key-based memoization behaves the way it would in a real
// component re-rendering from the same body.
func callFormatHelper(st *StateValue[int]) Signal[string] {
	return st.Format("n=%d")
}

func TestFormatDoesNotLeakSubscribersOnRerender(t *testing.T) {
	scope := newScope(context.Background(), nil)
	st := getOrCreateState(scope, "test:count", 0)

	var sig Signal[string]
	for i := 0; i < 50; i++ {
		sig = callFormatHelper(st)
	}
	if sig == nil {
		t.Fatal("expected Format to return a signal")
	}
	if n := internalSubCount(st); n != 1 {
		t.Fatalf("expected 1 internal subscriber after 50 Format calls from one site, got %d (leak)", n)
	}

	// Distinct call site must produce a distinct cache entry, so a
	// second subscription is expected there.
	_ = st.Format("other=%d")
	if n := internalSubCount(st); n != 2 {
		t.Fatalf("expected 2 subscribers after a second Format from a new site, got %d", n)
	}
}

func TestFormatReusesCachedSignal(t *testing.T) {
	scope := newScope(context.Background(), nil)
	st := getOrCreateState(scope, "test:count", 0)

	a := callFormatHelper(st)
	b := callFormatHelper(st)
	if a != b {
		t.Fatal("expected repeated Format from the same site to return the cached signal")
	}
}

func callDerivedHelper(st *StateValue[int]) Signal[int] {
	return Derived(st, func(n int) int { return n * 2 })
}

func TestDerivedDoesNotLeakSubscribersOnRerender(t *testing.T) {
	scope := newScope(context.Background(), nil)
	st := getOrCreateState(scope, "test:count", 0)

	for i := 0; i < 50; i++ {
		_ = callDerivedHelper(st)
	}
	if n := internalSubCount(st); n != 1 {
		t.Fatalf("expected 1 subscriber after 50 Derived calls from one site, got %d (leak)", n)
	}
}

type derivedListModel struct {
	Items []string
}

func callDerivedListHelper(st *StateValue[derivedListModel]) *StateValue[[]string] {
	return DerivedList(st, func(m derivedListModel) []string { return m.Items })
}

func TestDerivedListDoesNotLeakSubscribersOnRerender(t *testing.T) {
	scope := newScope(context.Background(), nil)
	st := getOrCreateState(scope, "test:model", derivedListModel{Items: []string{"a"}})

	for i := 0; i < 50; i++ {
		_ = callDerivedListHelper(st)
	}
	if n := internalSubCount(st); n != 1 {
		t.Fatalf("expected 1 subscriber after 50 DerivedList calls from one site, got %d (leak)", n)
	}

	// The derived state itself should also still propagate updates.
	derived := callDerivedListHelper(st)
	st.Update(func(m derivedListModel) derivedListModel {
		m.Items = append(m.Items, "b")
		return m
	})
	if got := derived.Get(); len(got) != 2 || got[1] != "b" {
		t.Fatalf("expected derived list to propagate updates, got %#v", got)
	}
}

func TestPersistedStateDoesNotLeakSubscribersOnRerender(t *testing.T) {
	scope := newScope(context.Background(), nil)
	dir := t.TempDir()

	var st *StateValue[int]
	for i := 0; i < 50; i++ {
		st = PersistedState(scope, "count", 0, PersistDir(dir))
	}
	if st == nil {
		t.Fatal("expected PersistedState to return a state")
	}
	if n := internalSubCount(st); n != 1 {
		t.Fatalf("expected 1 subscriber after 50 PersistedState calls, got %d (leak)", n)
	}
}

// siblingCounter is a Viewable whose Body calls State(s, c.Start). When
// two siblings share one parent scope, the State callerKey collides and
// the second instance inherits the first's value. With per-component
// child scopes, each instance holds its own state.
type siblingCounter struct {
	Start int
}

func (c siblingCounter) Body(s *Scope) View {
	count := State(s, c.Start)
	return Text(fmt.Sprintf("%d", count.Get()))
}

func TestSiblingComponentsHaveIsolatedState(t *testing.T) {
	root := Define(func(s *Scope) View {
		return VStack(siblingCounter{Start: 1}, siblingCounter{Start: 2})
	})
	scope := newScope(context.Background(), nil)
	comp := root.(*componentNode)
	view := comp.renderFn(scope)
	spec := toSpecWithScope(view.viewNode(), scope)

	if len(spec.Children) != 2 {
		t.Fatalf("expected 2 counters, got %d", len(spec.Children))
	}
	if spec.Children[0].TextValue != "1" {
		t.Fatalf("expected first counter %q, got %q", "1", spec.Children[0].TextValue)
	}
	if spec.Children[1].TextValue != "2" {
		t.Fatalf("expected second counter %q, got %q (sibling state leaked via shared scope)", "2", spec.Children[1].TextValue)
	}
}

func TestSiblingComponentStatePersistsAcrossRenders(t *testing.T) {
	root := Define(func(s *Scope) View {
		return VStack(siblingCounter{Start: 1}, siblingCounter{Start: 2})
	})
	scope := newScope(context.Background(), nil)
	comp := root.(*componentNode)

	// First render creates the per-sibling child scopes.
	_ = toSpecWithScope(comp.renderFn(scope).viewNode(), scope)

	// Mutate the first child's state directly through its memoized scope.
	// Children rendered via renderSlot sit one child-scope deep under the
	// parent (keyed by slot ID); there is no extra "body" layer because
	// renderSlot is the thing that creates the component's scope.
	firstChild := scope.childScopeFor("child:pos:0")
	firstChild.mu.Lock()
	var firstState *StateValue[int]
	for _, v := range firstChild.states {
		if st, ok := v.(*StateValue[int]); ok {
			firstState = st
			break
		}
	}
	firstChild.mu.Unlock()
	if firstState == nil {
		t.Fatal("expected first sibling's child scope to hold its State")
	}
	firstState.Set(99)

	// Second render should reuse the same child scopes, so the first
	// sibling keeps its mutated value and the second stays at its
	// initial value. A broken fix (fresh scopes every render) would
	// reset both.
	spec := toSpecWithScope(comp.renderFn(scope).viewNode(), scope)
	if spec.Children[0].TextValue != "99" {
		t.Fatalf("expected first sibling to persist mutated state (99), got %q", spec.Children[0].TextValue)
	}
	if spec.Children[1].TextValue != "2" {
		t.Fatalf("expected second sibling to retain isolated initial value (2), got %q", spec.Children[1].TextValue)
	}
}

func TestChildScopeForIsStableAndIsolated(t *testing.T) {
	parent := newScope(context.Background(), nil)

	a1 := parent.childScopeFor("a")
	a2 := parent.childScopeFor("a")
	b := parent.childScopeFor("b")

	if a1 != a2 {
		t.Fatal("expected childScopeFor to memoize by id")
	}
	if a1 == b {
		t.Fatal("expected different ids to produce different scopes")
	}
	if a1.parent != parent {
		t.Fatal("expected child scope to reference its parent")
	}
}

func TestDestroyTearsDownChildScopes(t *testing.T) {
	parent := newScope(context.Background(), nil)
	child := parent.childScopeFor("slot")

	parent.destroy()

	if child.ctx.Err() == nil {
		t.Fatal("expected child scope's context to be cancelled after parent.destroy()")
	}
}
