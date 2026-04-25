package gova

import (
	"context"
	"image/color"
	"os"
	"sync"

	"fyne.io/fyne/v2"
	fyneApp "fyne.io/fyne/v2/app"

	fyneBridge "github.com/nv404/gova/internal/fyne"
)

type AppConfig struct {
	Title  string
	Width  float32
	Height float32
	Theme  *Theme

	// Icon is the path to a PNG used as the window icon, the macOS dock
	// icon, and the Linux/Windows taskbar icon while the app is running.
	// Empty means "use the platform default." 256×256 or 512×512 PNGs
	// scale best across all platforms. Missing files are a silent no-op
	// so a forgotten asset does not crash the app.
	Icon string

	// IconBytes is an alternative to Icon when the icon is embedded at
	// compile time (e.g. via go:embed). If both are set, IconBytes wins.
	IconBytes []byte
}

const (
	defaultWidth  float32 = 400
	defaultHeight float32 = 600
)

func provideOverlays(s *Scope, a fyne.App, w fyne.Window, bridge *fyneBridge.FyneBridge) {
	fns := &overlayFuncs{
		showAlert: func(title, message string, actions []AlertAction) {
			specs := make([]fyneBridge.AlertActionSpec, len(actions))
			for i, act := range actions {
				specs[i] = fyneBridge.AlertActionSpec{
					Label:  act.Label,
					Style:  int(act.Style),
					Action: act.Action,
				}
			}
			fyne.Do(func() {
				fyneBridge.ShowAlert(w, title, message, specs)
			})
		},
		showSheet: func(content View, onDismiss func()) {
			spec := toSpec(content.viewNode())
			fyne.Do(func() {
				mounted := bridge.Mount(spec)
				fyneBridge.ShowSheet(w, mounted.Object, onDismiss)
			})
		},
		setTheme: func(t *Theme) {
			tc := fyneBridge.ThemeConfig{
				Variant: int(t.Variant),
				Accent:  t.Accent,
			}
			if t.Colors != nil {
				tc.Colors = make(map[int]color.Color, len(t.Colors))
				for k, v := range t.Colors {
					tc.Colors[int(k)] = v
				}
			}
			fyne.Do(func() {
				a.Settings().SetTheme(fyneBridge.NewTheme(tc))
			})
		},
	}
	Provide(s, overlayStoreKey, fns)
}

func Run(title string, root View) {
	RunWithConfig(AppConfig{Title: title}, root)
}

func RunWithConfig(config AppConfig, root View) {
	a := fyneApp.New()

	if config.Theme != nil {
		tc := fyneBridge.ThemeConfig{
			Variant: int(config.Theme.Variant),
			Accent:  config.Theme.Accent,
		}
		if config.Theme.Colors != nil {
			tc.Colors = make(map[int]color.Color, len(config.Theme.Colors))
			for k, v := range config.Theme.Colors {
				tc.Colors[int(k)] = v
			}
		}
		a.Settings().SetTheme(fyneBridge.NewTheme(tc))
	}

	iconBytes := loadIconBytes(config)
	if len(iconBytes) > 0 {
		res := fyne.NewStaticResource("gova-app-icon", iconBytes)
		a.SetIcon(res)
		setMacAppIcon(iconBytes)
	}

	w := a.NewWindow(config.Title)
	if len(iconBytes) > 0 {
		w.SetIcon(fyne.NewStaticResource("gova-app-icon", iconBytes))
	}

	width := config.Width
	height := config.Height
	if width <= 0 {
		width = defaultWidth
	}
	if height <= 0 {
		height = defaultHeight
	}
	w.Resize(fyne.NewSize(width, height))

	bridge := fyneBridge.New()

	comp, isComponent := root.(*componentNode)
	if !isComponent {
		spec := toSpec(root.viewNode())
		mounted := bridge.Mount(spec)
		w.SetContent(mounted.Object)
		w.ShowAndRun()
		return
	}

	var mu sync.Mutex
	var mounted *fyneBridge.MountedNode
	var scope *Scope

	scope = newScope(context.Background(), func() {
		mu.Lock()
		defer mu.Unlock()

		newView := renderComponent(comp, scope)
		newNode := newView.viewNode()
		comp.rendered = newNode
		newSpec := toSpecWithScope(newNode, scope)

		fyne.Do(func() {
			if mounted != nil {
				bridge.Reconcile(mounted, newSpec)
			}
		})
	})
	comp.scope = scope
	provideOverlays(scope, a, w, bridge)
	Provide[dialogPresenter](scope, dialogPresenterKey, newPlatformDialogPresenter(w))
	Provide(scope, themeStoreKey, config.Theme)

	firstView := renderComponent(comp, scope)
	firstNode := firstView.viewNode()
	comp.rendered = firstNode
	spec := toSpecWithScope(firstNode, scope)
	mounted = bridge.Mount(spec)
	w.SetContent(mounted.Object)

	w.ShowAndRun()
	mounted.Cleanup()
	scope.destroy()
}

// loadIconBytes resolves AppConfig.Icon / IconBytes into a byte slice.
// Returns nil on any error; a missing icon should never crash the app.
func loadIconBytes(cfg AppConfig) []byte {
	if len(cfg.IconBytes) > 0 {
		return cfg.IconBytes
	}
	if cfg.Icon == "" {
		return nil
	}
	data, err := os.ReadFile(cfg.Icon)
	if err != nil {
		return nil
	}
	return data
}
