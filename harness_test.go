package gova

import (
	"math"
	"sync/atomic"
	"testing"
	"time"
)

func TestTestRenderCounter(t *testing.T) {
	counter := Define(func(s *Scope) View {
		count := State(s, 0)
		return VStack(
			Text(count.Format("Count: %d")),
			Button("+", func() { count.Update(func(n int) int { return n + 1 }) }),
			Button("-", func() { count.Update(func(n int) int { return n - 1 }) }),
		)
	})

	tree := TestRender(counter)

	txt := tree.FindText(0)
	if !txt.Exists() {
		t.Fatal("expected to find text node")
	}
	if txt.Value != "Count: 0" {
		t.Fatalf("expected 'Count: 0', got %q", txt.Value)
	}

	tree.FindButton(0).Click()
	if tree.FindText(0).Value != "Count: 1" {
		t.Fatalf("expected 'Count: 1', got %q", tree.FindText(0).Value)
	}

	tree.FindButton(0).Click()
	tree.FindButton(0).Click()
	if tree.FindText(0).Value != "Count: 3" {
		t.Fatalf("expected 'Count: 3', got %q", tree.FindText(0).Value)
	}

	tree.FindButton(1).Click()
	if tree.FindText(0).Value != "Count: 2" {
		t.Fatalf("expected 'Count: 2', got %q", tree.FindText(0).Value)
	}
}

func TestTestRenderFindByLabel(t *testing.T) {
	app := Define(func(s *Scope) View {
		return VStack(
			Button("Save", func() {}),
			Button("Cancel", func() {}),
		)
	})

	tree := TestRender(app)

	save := tree.FindButtonByLabel("Save")
	if !save.Exists() {
		t.Fatal("expected to find Save button")
	}
	if save.Label != "Save" {
		t.Fatalf("expected label 'Save', got %q", save.Label)
	}

	missing := tree.FindButtonByLabel("Delete")
	if missing.Exists() {
		t.Fatal("expected Delete button to not exist")
	}
}

func TestTestRenderToggle(t *testing.T) {
	app := Define(func(s *Scope) View {
		enabled := State(s, false)
		return VStack(
			Toggle(enabled.Get()).Label("Feature").OnChange(func(v bool) { enabled.Set(v) }),
			When(enabled.Get(), func() View { return Text("ON") }),
		)
	})

	tree := TestRender(app)

	tog := tree.FindToggle(0)
	if !tog.Exists() {
		t.Fatal("expected toggle")
	}
	if tog.Checked {
		t.Fatal("expected unchecked initially")
	}

	if tree.NodeCount("Text") != 0 {
		t.Fatal("expected no text when toggle off")
	}

	tog.Toggle()
	if tree.NodeCount("Text") != 1 {
		t.Fatal("expected text after toggle on")
	}
	if tree.FindText(0).Value != "ON" {
		t.Fatalf("expected 'ON', got %q", tree.FindText(0).Value)
	}
}

func TestTestRenderTextField(t *testing.T) {
	app := Define(func(s *Scope) View {
		input := State(s, "")
		return VStack(
			TextField(input).Placeholder("Enter name"),
			Text(input.Format("Hello, %s")),
		)
	})

	tree := TestRender(app)

	tf := tree.FindTextField(0)
	if !tf.Exists() {
		t.Fatal("expected text field")
	}
	if tf.Placeholder != "Enter name" {
		t.Fatalf("expected placeholder, got %q", tf.Placeholder)
	}

	tf.Type("Alice")
	if tree.FindText(0).Value != "Hello, Alice" {
		t.Fatalf("expected 'Hello, Alice', got %q", tree.FindText(0).Value)
	}
}

func TestTestRenderWithStore(t *testing.T) {
	key := &StoreKey[string]{Default: "default"}

	inner := Define(func(s *Scope) View {
		store := UseStore(s, key)
		return Text(store.Get())
	})

	tree := TestRender(inner, WithStore(key, "injected"))

	if tree.FindText(0).Value != "injected" {
		t.Fatalf("expected 'injected', got %q", tree.FindText(0).Value)
	}
}

func TestTestRenderNodeCount(t *testing.T) {
	app := Define(func(s *Scope) View {
		return VStack(
			Text("a"),
			Text("b"),
			Button("x", func() {}),
			Divider(),
		)
	})

	tree := TestRender(app)

	if tree.NodeCount("Text") != 2 {
		t.Fatalf("expected 2 texts, got %d", tree.NodeCount("Text"))
	}
	if tree.NodeCount("Button") != 1 {
		t.Fatalf("expected 1 button, got %d", tree.NodeCount("Button"))
	}
	if tree.NodeCount("Divider") != 1 {
		t.Fatalf("expected 1 divider, got %d", tree.NodeCount("Divider"))
	}
}

func TestAnimationCurves(t *testing.T) {
	if applyCurve(0, EaseLinear) != 0 {
		t.Fatal("linear(0) should be 0")
	}
	if applyCurve(1, EaseLinear) != 1 {
		t.Fatal("linear(1) should be 1")
	}
	if applyCurve(0.5, EaseLinear) != 0.5 {
		t.Fatal("linear(0.5) should be 0.5")
	}

	if applyCurve(0, EaseInOut) != 0 {
		t.Fatal("easeInOut(0) should be 0")
	}
	if applyCurve(1, EaseInOut) != 1 {
		t.Fatal("easeInOut(1) should be 1")
	}

	mid := applyCurve(0.5, EaseInOut)
	if mid < 0.49 || mid > 0.51 {
		t.Fatalf("easeInOut(0.5) should be ~0.5, got %f", mid)
	}
}

func TestAnimationHandle(t *testing.T) {
	anim := Animate(50 * time.Millisecond)
	var lastProgress atomic.Uint32

	handle := &AnimationHandle{anim: anim}
	handle.Start(func(p float32) {
		lastProgress.Store(math.Float32bits(p))
	})

	time.Sleep(100 * time.Millisecond)

	progress := math.Float32frombits(lastProgress.Load())
	if !handle.Running() && progress < 0.9 {
		t.Fatalf("expected animation to complete, last progress: %f", progress)
	}

	time.Sleep(50 * time.Millisecond)
	if handle.Running() {
		t.Fatal("expected animation to stop after completion")
	}
}

func TestTestRenderNestedComponent(t *testing.T) {
	inner := Define(func(s *Scope) View {
		return Text("nested")
	})

	outer := Define(func(s *Scope) View {
		return VStack(Text("outer"), inner)
	})

	tree := TestRender(outer)

	if tree.NodeCount("Text") != 2 {
		t.Fatalf("expected 2 texts, got %d", tree.NodeCount("Text"))
	}
	if tree.FindText(1).Value != "nested" {
		t.Fatalf("expected 'nested', got %q", tree.FindText(1).Value)
	}
}
