//go:build darwin

package gova

/*
#cgo CFLAGS: -x objective-c -fobjc-arc
#cgo LDFLAGS: -framework AppKit -framework Foundation

#include <stdlib.h>
#import "dock_darwin.h"
*/
import "C"

import (
	"sync"
	"sync/atomic"
	"unsafe"
)

type darwinDock struct{}

func newPlatformDock() dockImpl { return darwinDock{} }

func (darwinDock) SetBadge(text string) {
	cstr := C.CString(text)
	defer C.free(unsafe.Pointer(cstr))
	C.govaDockSetBadge(cstr)
}

func (darwinDock) SetProgress(frac float64) { C.govaDockSetProgress(C.double(frac)) }

func (darwinDock) Bounce() { C.govaDockBounce() }

func (darwinDock) SetMenu(items []DockMenuItem) {
	count := len(items)
	if count == 0 {
		C.govaDockSetMenu(nil, nil, 0)
		return
	}
	cLabels := make([]*C.char, count)
	cIDs := make([]C.ulonglong, count)
	for i, it := range items {
		cLabels[i] = C.CString(it.Label)
		if it.Action != nil && it.Label != "" {
			cIDs[i] = C.ulonglong(registerDockHandler(it.Action))
		}
	}
	defer func() {
		for _, p := range cLabels {
			C.free(unsafe.Pointer(p))
		}
	}()
	C.govaDockSetMenu(
		(**C.char)(unsafe.Pointer(&cLabels[0])),
		(*C.ulonglong)(unsafe.Pointer(&cIDs[0])),
		C.int(count),
	)
}

func (darwinDock) Close() {}

// setMacAppIcon pushes the given PNG bytes to NSApp as the dock icon.
// Harmless when data is empty.
func setMacAppIcon(data []byte) {
	if len(data) == 0 {
		return
	}
	C.govaSetAppIcon(unsafe.Pointer(&data[0]), C.int(len(data)))
}

// Handler registry. Menu items cannot carry Go closures across the cgo
// boundary, so we keep them in a Go-side map keyed by a monotonic uint64 ID.
// The ObjC menu item stashes the ID in its NSMenuItem.tag; when the user
// clicks, ObjC calls govaDockMenuFire(id), which looks up and invokes the
// closure.
var (
	dockHandlerSeq atomic.Uint64
	dockHandlerMu  sync.Mutex
	dockHandlers   = map[uint64]func(){}
)

func registerDockHandler(fn func()) uint64 {
	id := dockHandlerSeq.Add(1)
	dockHandlerMu.Lock()
	dockHandlers[id] = fn
	dockHandlerMu.Unlock()
	return id
}

//export govaDockMenuFire
func govaDockMenuFire(id C.ulonglong) {
	dockHandlerMu.Lock()
	fn := dockHandlers[uint64(id)]
	dockHandlerMu.Unlock()
	if fn != nil {
		fn()
	}
}
