package fyneBridge

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
)

func ShowAlert(win fyne.Window, title, message string, actions []AlertActionSpec) {
	if len(actions) == 0 {
		dialog.ShowInformation(title, message, win)
		return
	}

	content := widget.NewLabel(message)
	content.Wrapping = fyne.TextWrapWord

	d := dialog.NewCustomWithoutButtons(title, content, win)

	buttons := make([]fyne.CanvasObject, len(actions))
	for i, act := range actions {
		a := act
		btn := widget.NewButton(a.Label, func() {
			d.Hide()
			if a.Action != nil {
				a.Action()
			}
		})
		if a.Style == ActionDestructive {
			btn.Importance = widget.DangerImportance
		} else if a.Style == ActionCancel {
			btn.Importance = widget.LowImportance
		} else {
			btn.Importance = widget.HighImportance
		}
		buttons[i] = btn
	}

	d.SetButtons(buttons)
	d.Resize(fyne.NewSize(400, 0))
	d.Show()
}

func ShowSheet(win fyne.Window, content fyne.CanvasObject, onDismiss func()) func() {
	wrapped := container.New(layout.NewCustomPaddedLayout(10, 10, 10, 10), content)
	d := dialog.NewCustom("", "Close", wrapped, win)
	d.SetOnClosed(onDismiss)
	d.Resize(fyne.NewSize(400, 300))
	d.Show()
	return d.Hide
}

type AlertActionSpec struct {
	Label  string
	Style  int // 0=default, 1=cancel, 2=destructive
	Action func()
}

const (
	ActionDefault     = 0
	ActionCancel      = 1
	ActionDestructive = 2
)
