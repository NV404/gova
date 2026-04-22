package main

import (
	"fmt"

	g "github.com/nv404/gova"
)

var App = g.Define(func(s *g.Scope) g.View {
	isDark := g.State(s, true)
	count := g.State(s, 0)
	showAlert := g.UseAlert(s)
	showSheet, _ := g.UseSheet(s)
	setTheme := g.UseSetTheme(s)

	return g.VStack(
		g.Text("Themed App").Font(g.Title).Bold(),
		g.Divider(),

		g.HStack(
			g.Text("Dark Mode"),
			g.Spacer(),
			g.Toggle(isDark.Get()).OnChange(func(dark bool) {
				isDark.Set(dark)
				if dark {
					setTheme(g.DarkTheme().WithAccent(g.Purple))
				} else {
					setTheme(g.LightTheme().WithAccent(g.Purple))
				}
			}),
		),

		g.Divider(),
		g.Text("Overlays").Font(g.Title2).Bold(),

		g.Button("Show Alert", func() {
			showAlert("Confirm Action", "Are you sure you want to proceed?",
				g.AlertAction{Label: "Cancel", Style: g.ActionCancel},
				g.AlertAction{Label: "Proceed", Style: g.ActionDefault, Action: func() {
					count.Update(func(n int) int { return n + 1 })
				}},
			)
		}),

		g.Button("Show Destructive Alert", func() {
			showAlert("Delete Everything", "This action cannot be undone.",
				g.AlertAction{Label: "Cancel", Style: g.ActionCancel},
				g.AlertAction{Label: "Delete", Style: g.ActionDestructive, Action: func() {
					count.Set(0)
				}},
			)
		}),

		g.Button("Show Sheet", func() {
			showSheet(
				g.VStack(
					g.Text("Sheet Content").Font(g.Title2).Bold(),
					g.Divider(),
					g.Text("This is a custom sheet overlay."),
					g.Text(fmt.Sprintf("Current count: %d", count.Get())),
				),
			)
		}),

		g.Divider(),
		g.Text(count.Format("Action count: %d")),

		g.Divider(),
		g.Text("Custom Colors").Font(g.Title2).Bold(),
		g.Text("Primary color").Color(g.Purple),
		g.Text("Error color").Color(g.Red),
		g.Text("Success color").Color(g.Green),
		g.Text("Warning color").Color(g.Orange),
		g.Text("Small caption").Font(g.Caption).Color(g.Gray),

		g.Spacer(),
	).Padding(20)
})

func main() {
	g.RunWithConfig(g.AppConfig{
		Title:  "Themed App",
		Width:  500,
		Height: 700,
		Theme:  g.DarkTheme().WithAccent(g.Purple),
	}, App)
}
