package gova

import (
	"context"
	"testing"
)

// fakeOverlay records calls and exercises the cancel-handle contract that
// the real Fyne-backed overlayFuncs implement. Each call to showSheet
// returns a fresh cancel closure; invoking it counts toward cancelCount
// and fires the onDismiss callback (mirroring Fyne's CustomDialog, which
// calls SetOnClosed on programmatic Hide).
type fakeOverlay struct {
	showCount   int
	cancelCount int
	lastOnDismiss func()
}

func (f *fakeOverlay) funcs() *overlayFuncs {
	return &overlayFuncs{
		showSheet: func(content View, onDismiss func()) func() {
			f.showCount++
			f.lastOnDismiss = onDismiss
			closed := false
			return func() {
				if closed {
					return
				}
				closed = true
				f.cancelCount++
				if onDismiss != nil {
					onDismiss()
				}
			}
		},
	}
}

func TestUseSheetDismissInvokesCancel(t *testing.T) {
	scope := newScope(context.Background(), nil)
	fake := &fakeOverlay{}
	Provide(scope, overlayStoreKey, fake.funcs())

	show, dismiss := UseSheet(scope)

	show(Text("first"))
	if fake.showCount != 1 {
		t.Fatalf("expected 1 showSheet call, got %d", fake.showCount)
	}

	dismiss()
	if fake.cancelCount != 1 {
		t.Fatalf("dismiss did not invoke cancel hook: cancelCount=%d (UseSheet must wire dismiss to the cancel closure returned by showSheet)", fake.cancelCount)
	}
}

func TestUseSheetDismissBeforeShowIsNoOp(t *testing.T) {
	scope := newScope(context.Background(), nil)
	fake := &fakeOverlay{}
	Provide(scope, overlayStoreKey, fake.funcs())

	_, dismiss := UseSheet(scope)
	// Calling dismiss before any show should not panic and should not
	// invoke any cancel hook (there is nothing to cancel).
	dismiss()
	if fake.cancelCount != 0 {
		t.Fatalf("dismiss before show should be a no-op; cancelCount=%d", fake.cancelCount)
	}
}

func TestUseSheetDismissIsIdempotent(t *testing.T) {
	scope := newScope(context.Background(), nil)
	fake := &fakeOverlay{}
	Provide(scope, overlayStoreKey, fake.funcs())

	show, dismiss := UseSheet(scope)
	show(Text("x"))
	dismiss()
	dismiss()
	dismiss()

	if fake.cancelCount != 1 {
		t.Fatalf("repeated dismiss should not re-cancel: cancelCount=%d", fake.cancelCount)
	}
}

func TestUseSheetUserCloseClearsHandle(t *testing.T) {
	// Simulates the user clicking the sheet's Close button: Fyne fires
	// the onDismiss callback directly, without going through the cancel
	// hook. A subsequent programmatic dismiss() must then be a no-op.
	scope := newScope(context.Background(), nil)
	fake := &fakeOverlay{}
	Provide(scope, overlayStoreKey, fake.funcs())

	show, dismiss := UseSheet(scope)
	show(Text("x"))

	if fake.lastOnDismiss == nil {
		t.Fatal("expected showSheet to receive an onDismiss callback")
	}
	fake.lastOnDismiss() // user closed the sheet themselves

	dismiss()
	if fake.cancelCount != 0 {
		t.Fatalf("dismiss after user-close should be no-op; cancelCount=%d", fake.cancelCount)
	}
}

func TestUseSheetReshowGetsFreshCancel(t *testing.T) {
	scope := newScope(context.Background(), nil)
	fake := &fakeOverlay{}
	Provide(scope, overlayStoreKey, fake.funcs())

	show, dismiss := UseSheet(scope)

	show(Text("first"))
	dismiss()
	if fake.cancelCount != 1 {
		t.Fatalf("first dismiss did not cancel: cancelCount=%d", fake.cancelCount)
	}

	show(Text("second"))
	if fake.showCount != 2 {
		t.Fatalf("expected 2 showSheet calls after reshow; got %d", fake.showCount)
	}
	dismiss()
	if fake.cancelCount != 2 {
		t.Fatalf("dismiss after reshow did not cancel new sheet: cancelCount=%d", fake.cancelCount)
	}
}
