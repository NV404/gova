package gova

import (
	"sync"
)

// Dock exposes the OS application dock (macOS), taskbar (Windows), or
// launcher entry (Linux desktops that implement the Unity launcher API).
//
// A Dock is one-per-process: call UseDock(s) from any component to get the
// shared handle. On platforms where an operation is not supported, the call
// is a silent no-op.
type Dock struct {
	mu   sync.Mutex
	impl dockImpl
}

// DockMenuItem is one entry in the dock right-click menu. An empty Label
// renders as a separator; an empty Action is disabled.
type DockMenuItem struct {
	Label  string
	Action func()
}

// dockImpl is the platform-specific surface. One implementation per GOOS,
// plus a no-op fallback.
type dockImpl interface {
	SetBadge(text string)
	SetProgress(fraction float64)
	Bounce()
	SetMenu(items []DockMenuItem)
	Close()
}

var (
	dockOnce   sync.Once
	dockShared *Dock
)

// sharedDock returns the process-wide Dock, creating it lazily.
func sharedDock() *Dock {
	dockOnce.Do(func() {
		dockShared = &Dock{impl: newPlatformDock()}
	})
	return dockShared
}

// UseDock returns the process-wide Dock handle. Scope is accepted so future
// versions can scope menus to a specific window; today it is unused.
func UseDock(_ *Scope) *Dock { return sharedDock() }

// SetBadge sets the dock badge text. Pass "" to clear.
func (d *Dock) SetBadge(text string) {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.impl.SetBadge(text)
}

// SetProgress sets the dock progress indicator. fraction is in [0, 1].
// Pass a negative value to hide the indicator.
func (d *Dock) SetProgress(fraction float64) {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.impl.SetProgress(fraction)
}

// Bounce requests the user's attention. On macOS this bounces the dock icon;
// on Windows it flashes the taskbar entry; on Linux it is a best-effort call.
func (d *Dock) Bounce() {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.impl.Bounce()
}

// SetMenu replaces the dock right-click menu. Pass nil to clear.
func (d *Dock) SetMenu(items []DockMenuItem) {
	d.mu.Lock()
	defer d.mu.Unlock()
	// Defensive copy so later caller mutations do not affect us.
	cpy := make([]DockMenuItem, len(items))
	copy(cpy, items)
	d.impl.SetMenu(cpy)
}

// setImpl swaps the underlying platform impl. Exposed for tests in this
// package only; never exported.
func (d *Dock) setImpl(impl dockImpl) {
	d.mu.Lock()
	defer d.mu.Unlock()
	if d.impl != nil {
		d.impl.Close()
	}
	d.impl = impl
}

// recordingDock is a dockImpl that records every call. It lets tests observe
// behavior without touching the real OS dock.
type recordingDock struct {
	Badges    []string
	Progress  []float64
	Bounces   int
	Menus     [][]DockMenuItem
	CloseHits int
}

func (r *recordingDock) SetBadge(t string)         { r.Badges = append(r.Badges, t) }
func (r *recordingDock) SetProgress(f float64)     { r.Progress = append(r.Progress, f) }
func (r *recordingDock) Bounce()                   { r.Bounces++ }
func (r *recordingDock) SetMenu(m []DockMenuItem) {
	cpy := make([]DockMenuItem, len(m))
	copy(cpy, m)
	r.Menus = append(r.Menus, cpy)
}
func (r *recordingDock) Close() { r.CloseHits++ }

// noopDock does nothing and returns nothing. Used as the fallback on
// platforms without a real implementation.
type noopDock struct{}

func (noopDock) SetBadge(string)          {}
func (noopDock) SetProgress(float64)      {}
func (noopDock) Bounce()                  {}
func (noopDock) SetMenu([]DockMenuItem)   {}
func (noopDock) Close()                   {}
