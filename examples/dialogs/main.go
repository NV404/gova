package main

import (
	"context"
	_ "embed"
	"os"
	"path/filepath"

	g "github.com/nv404/gova"
)

//go:embed icon.png
var appIcon []byte

var App = g.Define(func(s *g.Scope) g.View {
	status := g.State(s, "Pick a dialog to try.")
	dock := g.UseDock(s)

	// Dock menu is wired up once on first render so right-clicking the
	// app icon on macOS shows a shortcut to clear the status line.
	g.UseEffect(s, func(ctx context.Context) func() {
		dock.SetMenu([]g.DockMenuItem{
			{Label: "Reset status", Action: func() { status.Set("Status cleared from dock menu.") }},
			{Label: ""},
			{Label: "About", Action: func() { status.Set("Gova dialogs demo.") }},
		})
		return func() { dock.SetMenu(nil) }
	})

	return g.VStack(
		g.Text("Native dialogs").Font(g.Title).Bold(),
		g.Text("Each button presents a system-style dialog. The outcome appears below.").
			Color(g.Secondary),

		g.Divider(),

		g.Button("Show alert", func() {
			g.NativeAlert("Document saved", "Your draft has been saved to disk.").
				OK("Nice").
				OnClose(func() { status.Set("Alert dismissed.") }).
				Present(s)
		}),

		g.Button("Confirm (destructive)", func() {
			g.NativeConfirm("Delete this note?", "This cannot be undone.").
				Destructive().
				Confirm("Delete").
				Cancel("Keep").
				OnConfirm(func() {
					status.Set("Confirmed: note deleted.")
					dock.Bounce()
				}).
				OnCancel(func() { status.Set("Cancelled.") }).
				Present(s)
		}),

		g.Divider(),

		g.Button("Open file...", func() {
			g.FilePicker().
				Types(".png", ".jpg", ".jpeg", ".txt", ".md").
				StartDir(homeDir()).
				OnPick(func(path string) { status.Set("Opened: " + path) }).
				OnCancel(func() { status.Set("Open cancelled.") }).
				OnError(func(err error) { status.Set("Open error: " + err.Error()) }).
				Present(s)
		}),

		g.Button("Save file...", func() {
			g.SavePicker().
				Default("untitled.md").
				Types(".md", ".txt").
				StartDir(homeDir()).
				OnSave(func(path string) {
					_ = os.WriteFile(path, []byte("# Hello from Gova\n"), 0o644)
					status.Set("Wrote to " + path)
				}).
				OnCancel(func() { status.Set("Save cancelled.") }).
				OnError(func(err error) { status.Set("Save error: " + err.Error()) }).
				Present(s)
		}),

		g.Button("Pick folder...", func() {
			g.FolderPicker().
				StartDir(homeDir()).
				OnPick(func(path string) { status.Set("Selected folder: " + path) }).
				OnCancel(func() { status.Set("Folder pick cancelled.") }).
				Present(s)
		}),

		g.Divider(),

		g.HStack(
			g.Button("Badge '3'", func() { dock.SetBadge("3") }),
			g.Button("Clear badge", func() { dock.SetBadge("") }),
		).Spacing(g.SpaceSM),

		g.HStack(
			g.Button("Progress 40%", func() { dock.SetProgress(0.4) }),
			g.Button("Progress 100%", func() { dock.SetProgress(1.0) }),
			g.Button("Clear progress", func() { dock.SetProgress(-1) }),
		).Spacing(g.SpaceSM),

		g.Spacer(),

		g.Text(status.Format("%s")).
			Color(g.Secondary).
			Padding(g.SpaceSM),
	).
		Spacing(g.SpaceMD).
		Padding(g.SpaceLG)
})

// homeDir returns the user's home directory, falling back to "" so the
// picker picks its own sensible default.
func homeDir() string {
	h, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	return filepath.Clean(h)
}

func main() {
	g.RunWithConfig(g.AppConfig{
		Title:     "Gova Dialogs",
		Width:     460,
		Height:    620,
		IconBytes: appIcon,
	}, App)
}
