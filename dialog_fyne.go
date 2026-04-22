package gova

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/storage"
)

// fyneDialogPresenter is the Fyne-backed dialogPresenter published by
// RunWithConfig. All dialog methods dispatch to the Fyne main thread via
// fyne.Do so they can be called from any goroutine.
type fyneDialogPresenter struct {
	win fyne.Window
}

// newFyneDialogPresenter binds the presenter to a specific window. All
// dialogs are shown modal to this window.
func newFyneDialogPresenter(w fyne.Window) *fyneDialogPresenter {
	return &fyneDialogPresenter{win: w}
}

func (p *fyneDialogPresenter) ShowAlert(opts AlertOptions) {
	fyne.Do(func() {
		d := dialog.NewInformation(opts.Title, opts.Message, p.win)
		if opts.OK != "" {
			d.SetDismissText(opts.OK)
		}
		if opts.OnClose != nil {
			d.SetOnClosed(opts.OnClose)
		}
		d.Show()
	})
}

func (p *fyneDialogPresenter) ShowConfirm(opts ConfirmOptions) {
	fyne.Do(func() {
		d := dialog.NewConfirm(opts.Title, opts.Message, func(ok bool) {
			if ok {
				if opts.OnConfirm != nil {
					opts.OnConfirm()
				}
				return
			}
			if opts.OnCancel != nil {
				opts.OnCancel()
			}
		}, p.win)
		if opts.Confirm != "" {
			d.SetConfirmText(opts.Confirm)
		}
		if opts.Cancel != "" {
			d.SetDismissText(opts.Cancel)
		}
		if opts.Destructive {
			d.SetConfirmImportance(2) // widget.DangerImportance
		}
		d.Show()
	})
}

func (p *fyneDialogPresenter) ShowFileOpen(opts FilePickerOptions) {
	fyne.Do(func() {
		d := dialog.NewFileOpen(func(r fyne.URIReadCloser, err error) {
			if err != nil {
				if opts.OnError != nil {
					opts.OnError(err)
				}
				return
			}
			if r == nil {
				if opts.OnCancel != nil {
					opts.OnCancel()
				}
				return
			}
			path := r.URI().Path()
			_ = r.Close()
			if opts.OnPick != nil {
				opts.OnPick(path)
			}
		}, p.win)
		if len(opts.Extensions) > 0 {
			d.SetFilter(storage.NewExtensionFileFilter(opts.Extensions))
		}
		if opts.StartDir != "" {
			if u := listableURI(opts.StartDir); u != nil {
				d.SetLocation(u)
			}
		}
		d.Show()
	})
}

func (p *fyneDialogPresenter) ShowFileSave(opts SavePickerOptions) {
	fyne.Do(func() {
		d := dialog.NewFileSave(func(w fyne.URIWriteCloser, err error) {
			if err != nil {
				if opts.OnError != nil {
					opts.OnError(err)
				}
				return
			}
			if w == nil {
				if opts.OnCancel != nil {
					opts.OnCancel()
				}
				return
			}
			path := w.URI().Path()
			_ = w.Close()
			if opts.OnSave != nil {
				opts.OnSave(path)
			}
		}, p.win)
		if opts.DefaultName != "" {
			d.SetFileName(opts.DefaultName)
		}
		if len(opts.Extensions) > 0 {
			d.SetFilter(storage.NewExtensionFileFilter(opts.Extensions))
		}
		if opts.StartDir != "" {
			if u := listableURI(opts.StartDir); u != nil {
				d.SetLocation(u)
			}
		}
		d.Show()
	})
}

func (p *fyneDialogPresenter) ShowFolderOpen(opts FolderPickerOptions) {
	fyne.Do(func() {
		d := dialog.NewFolderOpen(func(l fyne.ListableURI, err error) {
			if err != nil {
				if opts.OnError != nil {
					opts.OnError(err)
				}
				return
			}
			if l == nil {
				if opts.OnCancel != nil {
					opts.OnCancel()
				}
				return
			}
			if opts.OnPick != nil {
				opts.OnPick(l.Path())
			}
		}, p.win)
		if opts.StartDir != "" {
			if u := listableURI(opts.StartDir); u != nil {
				d.SetLocation(u)
			}
		}
		d.Show()
	})
}

// listableURI converts a filesystem path to a fyne.ListableURI, returning nil
// if the conversion fails. Used when setting the start directory.
func listableURI(path string) fyne.ListableURI {
	u := storage.NewFileURI(path)
	lu, err := storage.ListerForURI(u)
	if err != nil {
		return nil
	}
	return lu
}
