package fyneBridge

import (
	"os"
	"testing"

	"fyne.io/fyne/v2/test"
	"fyne.io/fyne/v2/widget"
)

func TestMain(m *testing.M) {
	test.NewApp()
	os.Exit(m.Run())
}

// When a TextField is wrapped by min-size padding, the Entry must still be
// reachable via node.Inner so reconcile can update its text (e.g. clearing
// on submit). This guards the regression where Grow + MinHeight wrapping
// hid the Entry under a container.Stack.
func TestMountTextFieldWithMinSizePreservesInner(t *testing.T) {
	stateText := "hello"

	spec := ViewSpec{
		Kind:       KindTextField,
		HasMinSize: true,
		MinHeight:  36,
		TextFieldGet: func() string {
			return stateText
		},
		TextFieldSetSilent: func(s string) { stateText = s },
	}

	b := New()
	node := b.Mount(spec)

	entry, ok := node.Inner.(*widget.Entry)
	if !ok {
		t.Fatalf("expected node.Inner to be *widget.Entry after min-size wrap, got %T", node.Inner)
	}
	if entry.Text != "hello" {
		t.Fatalf("expected initial entry text %q, got %q", "hello", entry.Text)
	}

	// Simulate submit clearing state; reconcile must propagate to the entry.
	stateText = ""
	b.Reconcile(node, spec)

	entry2, ok := node.Inner.(*widget.Entry)
	if !ok {
		t.Fatalf("Inner lost Entry identity after reconcile: %T", node.Inner)
	}
	if entry2.Text != "" {
		t.Fatalf("expected entry text cleared after reconcile, got %q", entry2.Text)
	}
}

// A TextField with Grow set alongside HasMinSize should also preserve Inner.
func TestMountTextFieldWithGrowAndMinSizePreservesInner(t *testing.T) {
	stateText := "abc"
	spec := ViewSpec{
		Kind:         KindTextField,
		HasMinSize:   true,
		MinHeight:    36,
		Grow:         true,
		TextFieldGet: func() string { return stateText },
	}

	node := New().Mount(spec)

	if _, ok := node.Inner.(*widget.Entry); !ok {
		t.Fatalf("expected node.Inner to be *widget.Entry, got %T", node.Inner)
	}
}

// Background + padding wrapping should also preserve Inner for reconcile.
func TestMountWithBackgroundAndPaddingPreservesInner(t *testing.T) {
	spec := ViewSpec{
		Kind:            KindTextField,
		HasPadding:      true,
		PaddingTop:      8,
		PaddingBottom:   8,
		PaddingLeft:     8,
		PaddingRight:    8,
		HasBackground:   true,
		BackgroundColor: [4]uint8{0, 0, 0, 255},
		TextFieldGet:    func() string { return "" },
	}

	node := New().Mount(spec)
	if _, ok := node.Inner.(*widget.Entry); !ok {
		t.Fatalf("expected node.Inner to be *widget.Entry after background+padding wrap, got %T", node.Inner)
	}
}
