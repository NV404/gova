package gova

// This file defines the public fluent API for native system dialogs:
// NativeAlert, NativeConfirm, FilePicker, SavePicker, FolderPicker.

// dialogPresenterKey is the store key for the active presenter.
var dialogPresenterKey = &StoreKey[dialogPresenter]{}

// dialogPresenter is the interface the runtime provides to a Scope for showing OS-level dialogs.
type dialogPresenter interface {
	ShowAlert(opts AlertOptions)
	ShowConfirm(opts ConfirmOptions)
	ShowFileOpen(opts FilePickerOptions)
	ShowFileSave(opts SavePickerOptions)
	ShowFolderOpen(opts FolderPickerOptions)
}

// AlertOptions is the resolved payload for a NativeAlert.
type AlertOptions struct {
	Title   string
	Message string
	OK      string // defaults to "OK" when empty
	OnClose func()
}

// ConfirmOptions is the resolved payload for a NativeConfirm.
type ConfirmOptions struct {
	Title       string
	Message     string
	Confirm     string // defaults to "Confirm" when empty
	Cancel      string // defaults to "Cancel" when empty
	Destructive bool
	OnConfirm   func()
	OnCancel    func()
}

// FilePickerOptions is the resolved payload for a FilePicker.
type FilePickerOptions struct {
	Extensions []string // e.g. {".png", ".jpg"}; empty means all files
	StartDir   string
	OnPick     func(path string)
	OnCancel   func()
	OnError    func(err error)
}

// SavePickerOptions is the resolved payload for a SavePicker.
type SavePickerOptions struct {
	DefaultName string
	Extensions  []string
	StartDir    string
	OnSave      func(path string)
	OnCancel    func()
	OnError     func(err error)
}

// FolderPickerOptions is the resolved payload for a FolderPicker.
type FolderPickerOptions struct {
	StartDir string
	OnPick   func(path string)
	OnCancel func()
	OnError  func(err error)
}

// NativeAlertBuilder builds a single-button informational alert.
type NativeAlertBuilder struct{ opts AlertOptions }

// NativeAlert starts an alert builder. Call Present(s) to show it.
func NativeAlert(title, message string) *NativeAlertBuilder {
	return &NativeAlertBuilder{opts: AlertOptions{Title: title, Message: message}}
}

// OK overrides the confirm button label. Defaults to "OK".
func (b *NativeAlertBuilder) OK(label string) *NativeAlertBuilder {
	b.opts.OK = label
	return b
}

// OnClose sets a callback invoked after the user dismisses the alert.
func (b *NativeAlertBuilder) OnClose(fn func()) *NativeAlertBuilder {
	b.opts.OnClose = fn
	return b
}

// Present shows the alert using the presenter attached to s.
func (b *NativeAlertBuilder) Present(s *Scope) {
	if p := resolvePresenter(s); p != nil {
		p.ShowAlert(b.opts)
	}
}

// NativeConfirmBuilder builds a two-button confirmation dialog.
type NativeConfirmBuilder struct{ opts ConfirmOptions }

// NativeConfirm starts a confirm builder.
func NativeConfirm(title, message string) *NativeConfirmBuilder {
	return &NativeConfirmBuilder{opts: ConfirmOptions{Title: title, Message: message}}
}

// Confirm overrides the confirm button label.
func (b *NativeConfirmBuilder) Confirm(label string) *NativeConfirmBuilder {
	b.opts.Confirm = label
	return b
}

// Cancel overrides the cancel button label.
func (b *NativeConfirmBuilder) Cancel(label string) *NativeConfirmBuilder {
	b.opts.Cancel = label
	return b
}

// Destructive marks the confirm action as destructive. Presenters style it
// as a warning (red text on macOS, for example).
func (b *NativeConfirmBuilder) Destructive() *NativeConfirmBuilder {
	b.opts.Destructive = true
	return b
}

// OnConfirm is called when the user taps the confirm button.
func (b *NativeConfirmBuilder) OnConfirm(fn func()) *NativeConfirmBuilder {
	b.opts.OnConfirm = fn
	return b
}

// OnCancel is called when the user taps the cancel button.
func (b *NativeConfirmBuilder) OnCancel(fn func()) *NativeConfirmBuilder {
	b.opts.OnCancel = fn
	return b
}

// Present shows the confirm dialog.
func (b *NativeConfirmBuilder) Present(s *Scope) {
	if p := resolvePresenter(s); p != nil {
		p.ShowConfirm(b.opts)
	}
}

// FilePickerBuilder builds an open-file dialog.
type FilePickerBuilder struct{ opts FilePickerOptions }

// FilePicker starts an open-file builder.
func FilePicker() *FilePickerBuilder { return &FilePickerBuilder{} }

// Types restricts the picker to files with the given extensions. Extensions
// may be written with or without the leading dot: ".png" and "png" are both
// accepted.
func (b *FilePickerBuilder) Types(exts ...string) *FilePickerBuilder {
	b.opts.Extensions = normalizeExtensions(exts)
	return b
}

// StartDir sets the directory the picker opens in.
func (b *FilePickerBuilder) StartDir(path string) *FilePickerBuilder {
	b.opts.StartDir = path
	return b
}

// OnPick is called with the selected absolute path when the user confirms.
func (b *FilePickerBuilder) OnPick(fn func(path string)) *FilePickerBuilder {
	b.opts.OnPick = fn
	return b
}

// OnCancel is called when the user dismisses the picker without a choice.
func (b *FilePickerBuilder) OnCancel(fn func()) *FilePickerBuilder {
	b.opts.OnCancel = fn
	return b
}

// OnError is called when the dialog surfaces a system error.
func (b *FilePickerBuilder) OnError(fn func(err error)) *FilePickerBuilder {
	b.opts.OnError = fn
	return b
}

// Present shows the picker.
func (b *FilePickerBuilder) Present(s *Scope) {
	if p := resolvePresenter(s); p != nil {
		p.ShowFileOpen(b.opts)
	}
}

// SavePickerBuilder builds a save-file dialog.
type SavePickerBuilder struct{ opts SavePickerOptions }

// SavePicker starts a save-file builder.
func SavePicker() *SavePickerBuilder { return &SavePickerBuilder{} }

// Default pre-fills the filename field.
func (b *SavePickerBuilder) Default(name string) *SavePickerBuilder {
	b.opts.DefaultName = name
	return b
}

// Types restricts the visible files to the given extensions.
func (b *SavePickerBuilder) Types(exts ...string) *SavePickerBuilder {
	b.opts.Extensions = normalizeExtensions(exts)
	return b
}

// StartDir sets the directory the picker opens in.
func (b *SavePickerBuilder) StartDir(path string) *SavePickerBuilder {
	b.opts.StartDir = path
	return b
}

// OnSave is called with the selected destination path.
func (b *SavePickerBuilder) OnSave(fn func(path string)) *SavePickerBuilder {
	b.opts.OnSave = fn
	return b
}

// OnCancel is called when the user cancels.
func (b *SavePickerBuilder) OnCancel(fn func()) *SavePickerBuilder {
	b.opts.OnCancel = fn
	return b
}

// OnError is called when a system error occurs.
func (b *SavePickerBuilder) OnError(fn func(err error)) *SavePickerBuilder {
	b.opts.OnError = fn
	return b
}

// Present shows the picker.
func (b *SavePickerBuilder) Present(s *Scope) {
	if p := resolvePresenter(s); p != nil {
		p.ShowFileSave(b.opts)
	}
}

// FolderPickerBuilder builds an open-folder dialog.
type FolderPickerBuilder struct{ opts FolderPickerOptions }

// FolderPicker starts a folder-picker builder.
func FolderPicker() *FolderPickerBuilder { return &FolderPickerBuilder{} }

// StartDir sets the directory the picker opens in.
func (b *FolderPickerBuilder) StartDir(path string) *FolderPickerBuilder {
	b.opts.StartDir = path
	return b
}

// OnPick is called with the selected folder path.
func (b *FolderPickerBuilder) OnPick(fn func(path string)) *FolderPickerBuilder {
	b.opts.OnPick = fn
	return b
}

// OnCancel is called when the user cancels.
func (b *FolderPickerBuilder) OnCancel(fn func()) *FolderPickerBuilder {
	b.opts.OnCancel = fn
	return b
}

// OnError is called when a system error occurs.
func (b *FolderPickerBuilder) OnError(fn func(err error)) *FolderPickerBuilder {
	b.opts.OnError = fn
	return b
}

// Present shows the picker.
func (b *FolderPickerBuilder) Present(s *Scope) {
	if p := resolvePresenter(s); p != nil {
		p.ShowFolderOpen(b.opts)
	}
}

// resolvePresenter returns the presenter published on s (or any ancestor), or nil if none has been published.
func resolvePresenter(s *Scope) dialogPresenter {
	if s == nil {
		return nil
	}
	store := UseStore(s, dialogPresenterKey)
	if store == nil {
		return nil
	}
	return store.Get()
}

// normalizeExtensions trims and lowercases each extension and ensures it
// starts with a dot.
func normalizeExtensions(in []string) []string {
	if len(in) == 0 {
		return nil
	}
	out := make([]string, 0, len(in))
	for _, s := range in {
		if s == "" {
			continue
		}
		// strings.ToLower without importing strings: manual cast.
		b := []byte(s)
		for i, c := range b {
			if c >= 'A' && c <= 'Z' {
				b[i] = c + 32
			}
		}
		s = string(b)
		if s[0] != '.' {
			s = "." + s
		}
		out = append(out, s)
	}
	return out
}
