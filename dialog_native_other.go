//go:build !darwin

package gova

import "fyne.io/fyne/v2"

// newPlatformDialogPresenter falls back to the Fyne-drawn presenter on
// platforms where we do not yet have a native implementation. Windows and
// Linux will grow real bridges later; until then users see a Fyne popup.
func newPlatformDialogPresenter(w fyne.Window) dialogPresenter {
	return newFyneDialogPresenter(w)
}
