//go:build !darwin

package gova

// newPlatformDock returns a no-op Dock on platforms without an implementation.
func newPlatformDock() dockImpl { return noopDock{} }

// setMacAppIcon is a no-op off macOS. The Windows taskbar / Linux WM icon is
// still driven by AppConfig.Icon through Fyne's SetIcon path.
func setMacAppIcon(_ []byte) {}
