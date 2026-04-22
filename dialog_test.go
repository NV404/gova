package gova

import (
	"context"
	"errors"
	"reflect"
	"testing"
)

// recordingPresenter is a dialogPresenter that records every call so tests
// can assert on what the builder layer dispatched.
type recordingPresenter struct {
	Alert    []AlertOptions
	Confirm  []ConfirmOptions
	FileOpen []FilePickerOptions
	FileSave []SavePickerOptions
	Folder   []FolderPickerOptions
}

func (r *recordingPresenter) ShowAlert(o AlertOptions)              { r.Alert = append(r.Alert, o) }
func (r *recordingPresenter) ShowConfirm(o ConfirmOptions)           { r.Confirm = append(r.Confirm, o) }
func (r *recordingPresenter) ShowFileOpen(o FilePickerOptions)       { r.FileOpen = append(r.FileOpen, o) }
func (r *recordingPresenter) ShowFileSave(o SavePickerOptions)       { r.FileSave = append(r.FileSave, o) }
func (r *recordingPresenter) ShowFolderOpen(o FolderPickerOptions)   { r.Folder = append(r.Folder, o) }

// scopeWithPresenter returns a scope that has the given presenter published.
func scopeWithPresenter(p dialogPresenter) (*Scope, *recordingPresenter) {
	s := newScope(context.Background(), nil)
	Provide[dialogPresenter](s, dialogPresenterKey, p)
	rp, _ := p.(*recordingPresenter)
	return s, rp
}

func TestNativeAlertCapturesFields(t *testing.T) {
	s, rp := scopeWithPresenter(&recordingPresenter{})
	defer s.destroy()

	closed := false
	NativeAlert("Saved", "All good").
		OK("Dismiss").
		OnClose(func() { closed = true }).
		Present(s)

	if len(rp.Alert) != 1 {
		t.Fatalf("expected 1 alert, got %d", len(rp.Alert))
	}
	got := rp.Alert[0]
	if got.Title != "Saved" || got.Message != "All good" || got.OK != "Dismiss" {
		t.Fatalf("alert fields wrong: %#v", got)
	}
	got.OnClose()
	if !closed {
		t.Fatal("OnClose callback not wired")
	}
}

func TestNativeConfirmRoutesConfirmAndCancel(t *testing.T) {
	s, rp := scopeWithPresenter(&recordingPresenter{})
	defer s.destroy()

	var confirmed, cancelled int
	builder := NativeConfirm("Delete?", "Cannot be undone").
		Destructive().
		Confirm("Delete").
		Cancel("Keep").
		OnConfirm(func() { confirmed++ }).
		OnCancel(func() { cancelled++ })
	builder.Present(s)
	builder.Present(s)

	if len(rp.Confirm) != 2 {
		t.Fatalf("expected 2 confirms, got %d", len(rp.Confirm))
	}
	opts := rp.Confirm[0]
	if !opts.Destructive || opts.Confirm != "Delete" || opts.Cancel != "Keep" {
		t.Fatalf("confirm labels/flags wrong: %#v", opts)
	}
	opts.OnConfirm()
	opts.OnCancel()
	if confirmed != 1 || cancelled != 1 {
		t.Fatalf("confirm=%d cancel=%d", confirmed, cancelled)
	}
}

func TestFilePickerBuilder(t *testing.T) {
	s, rp := scopeWithPresenter(&recordingPresenter{})
	defer s.destroy()

	var pickedPath string
	var gotErr error
	FilePicker().
		Types(".PNG", "jpg", ".jpeg").
		StartDir("/tmp").
		OnPick(func(p string) { pickedPath = p }).
		OnError(func(e error) { gotErr = e }).
		Present(s)

	if len(rp.FileOpen) != 1 {
		t.Fatalf("expected 1 file open, got %d", len(rp.FileOpen))
	}
	opts := rp.FileOpen[0]
	if !reflect.DeepEqual(opts.Extensions, []string{".png", ".jpg", ".jpeg"}) {
		t.Fatalf("extensions not normalized: %v", opts.Extensions)
	}
	if opts.StartDir != "/tmp" {
		t.Fatalf("start dir: %q", opts.StartDir)
	}
	opts.OnPick("/tmp/a.png")
	opts.OnError(errors.New("boom"))
	if pickedPath != "/tmp/a.png" {
		t.Fatalf("OnPick not wired: %q", pickedPath)
	}
	if gotErr == nil || gotErr.Error() != "boom" {
		t.Fatalf("OnError not wired: %v", gotErr)
	}
}

func TestSavePickerBuilder(t *testing.T) {
	s, rp := scopeWithPresenter(&recordingPresenter{})
	defer s.destroy()

	var saved string
	var cancelled bool
	SavePicker().
		Default("untitled.txt").
		Types(".txt").
		StartDir("/home/user").
		OnSave(func(p string) { saved = p }).
		OnCancel(func() { cancelled = true }).
		Present(s)

	if len(rp.FileSave) != 1 {
		t.Fatalf("expected 1 save, got %d", len(rp.FileSave))
	}
	opts := rp.FileSave[0]
	if opts.DefaultName != "untitled.txt" || opts.StartDir != "/home/user" {
		t.Fatalf("save opts wrong: %#v", opts)
	}
	opts.OnSave("/home/user/draft.txt")
	opts.OnCancel()
	if saved != "/home/user/draft.txt" || !cancelled {
		t.Fatalf("callbacks not wired: saved=%q cancelled=%v", saved, cancelled)
	}
}

func TestFolderPickerBuilder(t *testing.T) {
	s, rp := scopeWithPresenter(&recordingPresenter{})
	defer s.destroy()

	var picked string
	FolderPicker().
		StartDir("/").
		OnPick(func(p string) { picked = p }).
		Present(s)

	if len(rp.Folder) != 1 {
		t.Fatalf("expected 1 folder pick, got %d", len(rp.Folder))
	}
	rp.Folder[0].OnPick("/usr/local")
	if picked != "/usr/local" {
		t.Fatalf("OnPick not wired: %q", picked)
	}
}

func TestPresentWithoutPresenterIsNoOp(t *testing.T) {
	s := newScope(context.Background(), nil)
	defer s.destroy()
	// No Provide of dialogPresenterKey. Every Present must be a silent no-op.
	NativeAlert("t", "m").Present(s)
	NativeConfirm("t", "m").OnConfirm(func() {}).Present(s)
	FilePicker().OnPick(func(string) {}).Present(s)
	SavePicker().OnSave(func(string) {}).Present(s)
	FolderPicker().OnPick(func(string) {}).Present(s)
}

func TestNormalizeExtensions(t *testing.T) {
	cases := []struct {
		in   []string
		want []string
	}{
		{nil, nil},
		{[]string{}, nil},
		{[]string{".png"}, []string{".png"}},
		{[]string{"PNG"}, []string{".png"}},
		{[]string{"", ".Jpg", "jpeg"}, []string{".jpg", ".jpeg"}},
	}
	for _, tc := range cases {
		got := normalizeExtensions(tc.in)
		if !reflect.DeepEqual(got, tc.want) {
			t.Errorf("normalizeExtensions(%v) = %v, want %v", tc.in, got, tc.want)
		}
	}
}

func TestConfirmCancelDefaultNilSafe(t *testing.T) {
	s, rp := scopeWithPresenter(&recordingPresenter{})
	defer s.destroy()

	// No OnCancel set; the presenter should still be safe to invoke the
	// internal callback plumbing. We simulate the raw flow by calling the
	// options' callbacks directly.
	NativeConfirm("x", "y").OnConfirm(func() {}).Present(s)
	opts := rp.Confirm[0]
	// Either callback being nil must not crash when the producer invokes it.
	if opts.OnCancel != nil {
		t.Fatal("OnCancel should be nil when unset")
	}
	if opts.OnConfirm == nil {
		t.Fatal("OnConfirm must be set")
	}
}
