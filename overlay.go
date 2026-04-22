package gova

type ActionStyle int

const (
	ActionDefault     ActionStyle = iota
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
	showSheet func(content View, onDismiss func())
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
func UseSheet(s *Scope) (show func(content View), dismiss func()) {
	store := UseStore(s, overlayStoreKey)
	fns := store.Get()

	var dismissFn func()
	show = func(content View) {
		if fns != nil && fns.showSheet != nil {
			fns.showSheet(content, func() { dismissFn = nil })
		}
	}
	dismiss = func() {
		if dismissFn != nil {
			dismissFn()
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
