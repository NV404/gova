//go:build darwin

package gova

/*
#cgo CFLAGS: -x objective-c -fobjc-arc
#cgo LDFLAGS: -framework AppKit -framework Foundation

#include <stdlib.h>
#import "dialog_darwin.h"
*/
import "C"

import (
	"sync"
	"sync/atomic"
	"unsafe"

	"fyne.io/fyne/v2"
)

// nativeDialogPresenter shows true macOS system dialogs via NSAlert,
// NSOpenPanel and NSSavePanel. The fyne.Window is kept only to keep the
// constructor signature aligned with the Fyne fallback; the actual parent
// window is discovered at present time through [NSApp keyWindow].
type nativeDialogPresenter struct {
	_ fyne.Window
}

func newPlatformDialogPresenter(w fyne.Window) dialogPresenter {
	return &nativeDialogPresenter{}
}

// Completion-handler registry. Each Show* call allocates a single callbacks
// struct and passes its atomic ID across cgo; ObjC calls the matching
// exported Go function, which looks up the callbacks and clears the entry.
type dialogCallbacks struct {
	onAlert   func()
	onConfirm func(bool)
	onPath    func(path string, cancelled bool)
}

var (
	dialogCBSeq atomic.Uint64
	dialogCBMu  sync.Mutex
	dialogCBs   = map[uint64]*dialogCallbacks{}
)

func registerDialogCB(cb *dialogCallbacks) uint64 {
	id := dialogCBSeq.Add(1)
	dialogCBMu.Lock()
	dialogCBs[id] = cb
	dialogCBMu.Unlock()
	return id
}

func takeDialogCB(id uint64) *dialogCallbacks {
	dialogCBMu.Lock()
	cb := dialogCBs[id]
	delete(dialogCBs, id)
	dialogCBMu.Unlock()
	return cb
}

//export govaDialogAlertDone
func govaDialogAlertDone(cbid C.ulonglong) {
	cb := takeDialogCB(uint64(cbid))
	if cb != nil && cb.onAlert != nil {
		cb.onAlert()
	}
}

//export govaDialogConfirmDone
func govaDialogConfirmDone(cbid C.ulonglong, confirmed C.int) {
	cb := takeDialogCB(uint64(cbid))
	if cb != nil && cb.onConfirm != nil {
		cb.onConfirm(confirmed != 0)
	}
}

//export govaDialogPathDone
func govaDialogPathDone(cbid C.ulonglong, path *C.char, cancelled C.int) {
	cb := takeDialogCB(uint64(cbid))
	if cb == nil || cb.onPath == nil {
		return
	}
	p := ""
	if path != nil {
		p = C.GoString(path)
	}
	cb.onPath(p, cancelled != 0)
}

// cExts converts Go extension strings into a NULL-terminated C array; the
// returned free function releases every allocation.
func cExts(exts []string) (arr **C.char, count C.int, free func()) {
	if len(exts) == 0 {
		return nil, 0, func() {}
	}
	cstrs := make([]*C.char, len(exts))
	for i, e := range exts {
		cstrs[i] = C.CString(e)
	}
	return (**C.char)(unsafe.Pointer(&cstrs[0])), C.int(len(cstrs)), func() {
		for _, p := range cstrs {
			C.free(unsafe.Pointer(p))
		}
	}
}

func (p *nativeDialogPresenter) ShowAlert(opts AlertOptions) {
	id := registerDialogCB(&dialogCallbacks{onAlert: opts.OnClose})
	cTitle := C.CString(opts.Title)
	cMsg := C.CString(opts.Message)
	cOK := C.CString(opts.OK)
	defer C.free(unsafe.Pointer(cTitle))
	defer C.free(unsafe.Pointer(cMsg))
	defer C.free(unsafe.Pointer(cOK))
	C.govaShowAlert(cTitle, cMsg, cOK, C.ulonglong(id))
}

func (p *nativeDialogPresenter) ShowConfirm(opts ConfirmOptions) {
	id := registerDialogCB(&dialogCallbacks{onConfirm: func(ok bool) {
		if ok {
			if opts.OnConfirm != nil {
				opts.OnConfirm()
			}
			return
		}
		if opts.OnCancel != nil {
			opts.OnCancel()
		}
	}})
	cTitle := C.CString(opts.Title)
	cMsg := C.CString(opts.Message)
	cConfirm := C.CString(opts.Confirm)
	cCancel := C.CString(opts.Cancel)
	defer C.free(unsafe.Pointer(cTitle))
	defer C.free(unsafe.Pointer(cMsg))
	defer C.free(unsafe.Pointer(cConfirm))
	defer C.free(unsafe.Pointer(cCancel))
	destr := C.int(0)
	if opts.Destructive {
		destr = 1
	}
	C.govaShowConfirm(cTitle, cMsg, cConfirm, cCancel, destr, C.ulonglong(id))
}

func (p *nativeDialogPresenter) ShowFileOpen(opts FilePickerOptions) {
	id := registerDialogCB(&dialogCallbacks{onPath: func(path string, cancelled bool) {
		if cancelled {
			if opts.OnCancel != nil {
				opts.OnCancel()
			}
			return
		}
		if opts.OnPick != nil {
			opts.OnPick(path)
		}
	}})
	cStart := C.CString(opts.StartDir)
	defer C.free(unsafe.Pointer(cStart))
	exts, n, freeExts := cExts(opts.Extensions)
	defer freeExts()
	C.govaShowFileOpen(cStart, exts, n, C.ulonglong(id))
}

func (p *nativeDialogPresenter) ShowFileSave(opts SavePickerOptions) {
	id := registerDialogCB(&dialogCallbacks{onPath: func(path string, cancelled bool) {
		if cancelled {
			if opts.OnCancel != nil {
				opts.OnCancel()
			}
			return
		}
		if opts.OnSave != nil {
			opts.OnSave(path)
		}
	}})
	cStart := C.CString(opts.StartDir)
	cName := C.CString(opts.DefaultName)
	defer C.free(unsafe.Pointer(cStart))
	defer C.free(unsafe.Pointer(cName))
	exts, n, freeExts := cExts(opts.Extensions)
	defer freeExts()
	C.govaShowFileSave(cStart, cName, exts, n, C.ulonglong(id))
}

func (p *nativeDialogPresenter) ShowFolderOpen(opts FolderPickerOptions) {
	id := registerDialogCB(&dialogCallbacks{onPath: func(path string, cancelled bool) {
		if cancelled {
			if opts.OnCancel != nil {
				opts.OnCancel()
			}
			return
		}
		if opts.OnPick != nil {
			opts.OnPick(path)
		}
	}})
	cStart := C.CString(opts.StartDir)
	defer C.free(unsafe.Pointer(cStart))
	C.govaShowFolderOpen(cStart, C.ulonglong(id))
}
