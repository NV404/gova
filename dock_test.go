package gova

import (
	"testing"
)

func TestDockRecordsCallsThroughFake(t *testing.T) {
	d := &Dock{impl: &recordingDock{}}
	d.SetBadge("3")
	d.SetBadge("")
	d.SetProgress(0.5)
	d.SetProgress(-1)
	d.Bounce()
	d.Bounce()
	d.SetMenu([]DockMenuItem{{Label: "Quit", Action: func() {}}})

	r := d.impl.(*recordingDock)
	if got, want := r.Badges, []string{"3", ""}; !eqStrings(got, want) {
		t.Fatalf("Badges: got %v want %v", got, want)
	}
	if got, want := r.Progress, []float64{0.5, -1}; !eqFloats(got, want) {
		t.Fatalf("Progress: got %v want %v", got, want)
	}
	if r.Bounces != 2 {
		t.Fatalf("Bounces: got %d want 2", r.Bounces)
	}
	if len(r.Menus) != 1 || len(r.Menus[0]) != 1 || r.Menus[0][0].Label != "Quit" {
		t.Fatalf("Menus snapshot wrong: %#v", r.Menus)
	}
}

func TestDockSetMenuDefensiveCopy(t *testing.T) {
	d := &Dock{impl: &recordingDock{}}
	items := []DockMenuItem{{Label: "A"}, {Label: "B"}}
	d.SetMenu(items)
	items[0].Label = "Z"
	r := d.impl.(*recordingDock)
	if r.Menus[0][0].Label != "A" {
		t.Fatalf("SetMenu did not defensively copy: %q", r.Menus[0][0].Label)
	}
}

func TestSharedDockSingleton(t *testing.T) {
	a := sharedDock()
	b := sharedDock()
	if a != b {
		t.Fatal("sharedDock must return the same instance")
	}
	// UseDock must return the same handle regardless of scope.
	s := newScope(t.Context(), nil)
	defer s.destroy()
	if UseDock(s) != a {
		t.Fatal("UseDock must return the shared dock")
	}
}

func TestDockSetImplClosesPrior(t *testing.T) {
	first := &recordingDock{}
	d := &Dock{impl: first}
	second := &recordingDock{}
	d.setImpl(second)
	if first.CloseHits != 1 {
		t.Fatalf("prior impl must be Closed once, got %d", first.CloseHits)
	}
	d.SetBadge("x")
	if len(first.Badges) != 0 {
		t.Fatal("calls after setImpl must not reach prior impl")
	}
	if len(second.Badges) != 1 || second.Badges[0] != "x" {
		t.Fatalf("new impl should have received the call, got %v", second.Badges)
	}
}

func TestNoopDockIsInert(t *testing.T) {
	n := noopDock{}
	// No panics, no return values.
	n.SetBadge("hi")
	n.SetProgress(0.3)
	n.Bounce()
	n.SetMenu(nil)
	n.Close()
}

func eqStrings(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func eqFloats(a, b []float64) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
