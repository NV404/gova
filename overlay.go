package gova

import "sync"

type ActionStyle int

const (
	ActionDefault ActionStyle = iota
	ActionCancel
	ActionDestructive
)

type AlertAction struct {
	Label  string
	Style  ActionStyle
	Action func()
}

// showAlertFn / showSheetFn are set by RunWithConfig to bind the fyne.Window.
// Stored on the scope via an internal store so overlay hooks can access them.
var overlayStoreKey = &StoreKey[*overlayFuncs]{Default: nil}

type overlayFuncs struct {
	showAlert func(title, message string, actions []AlertAction)
	showSheet func(content View, onDismiss func()) func()
	setTheme  func(theme *Theme)
}

// UseAlert returns a function that shows a modal alert dialog.
// Call it from event handlers (button taps, etc).
func UseAlert(s *Scope) func(title, message string, actions ...AlertAction) {
	store := UseStore(s, overlayStoreKey)
	fns := store.Get()
	return func(title, message string, actions ...AlertAction) {
		if fns != nil && fns.showAlert != nil {
			fns.showAlert(title, message, actions)
		}
	}
}

// UseSheet returns (show, dismiss) functions for presenting a sheet overlay.
// dismiss programmatically closes the most recently shown sheet; it is a
// no-op if no sheet is currently presented (either never shown, already
// dismissed by the user, or already dismissed programmatically).
func UseSheet(s *Scope) (show func(content View), dismiss func()) {
	store := UseStore(s, overlayStoreKey)
	fns := store.Get()

	var (
		mu     sync.Mutex
		cancel func()
	)
	show = func(content View) {
		if fns == nil || fns.showSheet == nil {
			return
		}
		// onDismiss fires when the sheet is closed by any path
		// (user-driven Close button or programmatic cancel) so the
		// next dismiss() call doesn't try to cancel a dead sheet.
		c := fns.showSheet(content, func() {
			mu.Lock()
			cancel = nil
			mu.Unlock()
		})
		mu.Lock()
		cancel = c
		mu.Unlock()
	}
	dismiss = func() {
		mu.Lock()
		c := cancel
		cancel = nil
		mu.Unlock()
		if c != nil {
			c()
		}
	}
	return show, dismiss
}

// UseSetTheme returns a function that switches the app theme at runtime.
func UseSetTheme(s *Scope) func(*Theme) {
	store := UseStore(s, overlayStoreKey)
	fns := store.Get()
	return func(t *Theme) {
		if fns != nil && fns.setTheme != nil {
			fns.setTheme(t)
		}
	}
}
