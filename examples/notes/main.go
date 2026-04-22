package main

import (
	"fmt"

	g "github.com/nv404/gova"
)

type Note struct {
	ID    int
	Title string
	Body  string
}

type NotesModel struct {
	Notes  []Note
	NextID int
}

func addNote(m NotesModel, title string) NotesModel {
	m.Notes = append(m.Notes, Note{ID: m.NextID, Title: title})
	m.NextID++
	return m
}

func removeNote(m NotesModel, id int) NotesModel {
	for i := range m.Notes {
		if m.Notes[i].ID == id {
			m.Notes = append(m.Notes[:i], m.Notes[i+1:]...)
			return m
		}
	}
	return m
}

var NotesStore = &g.StoreKey[NotesModel]{Default: NotesModel{NextID: 1}}

var NotesTab = g.Define(func(s *g.Scope) g.View {
	model := g.UseStore(s, NotesStore)
	input := g.State(s, "")

	add := func() {
		title := input.Get()
		if title == "" {
			return
		}
		model.Update(func(m NotesModel) NotesModel {
			return addNote(m, title)
		})
		input.Set("")
	}

	notes := g.DerivedList(model, func(m NotesModel) []Note { return m.Notes })
	count := g.Derived(model, func(m NotesModel) string {
		return fmt.Sprintf("%d notes", len(m.Notes))
	})

	return g.Scaffold(
		g.ScrollView(
			g.List(notes,
				func(n Note) int { return n.ID },
				func(i int, note Note) g.View {
					id := note.ID
					return g.HStack(
						g.Text(note.Title),
						g.Spacer(),
						g.Button("Delete", func() {
							model.Update(func(m NotesModel) NotesModel {
								return removeNote(m, id)
							})
						}).Color(g.Red),
					)
				},
			),
		),
	).Top(
		g.VStack(
			g.Text("Notes").Font(g.Title).Bold(),
			g.HStack(
				g.TextField(input).
					Placeholder("New note title...").
					OnSubmit(func(val string) { add() }).
					MinHeight(36).
					Grow(),
				g.Button("Add", add),
			).Spacing(g.SpaceSM),
		).Spacing(g.SpaceMD).Padding(g.SpaceMD),
	).Bottom(
		g.Text(count).Font(g.Caption).Color(g.Secondary).Padding(g.SpaceMD),
	)
})

var StatsTab = g.Define(func(s *g.Scope) g.View {
	model := g.UseStore(s, NotesStore)

	count := g.Derived(model, func(m NotesModel) string {
		return fmt.Sprintf("Total notes: %d", len(m.Notes))
	})

	return g.VStack(
		g.Text("Statistics").Font(g.Title).Bold(),
		g.Divider(),
		g.Text(count),
		g.Spacer(),
	).Padding(16)
})

var ErrorTab = g.Define(func(s *g.Scope) g.View {
	errMsg := g.State(s, "")

	return g.VStack(
		g.Text("Error Boundary Demo").Font(g.Title).Bold(),
		g.Divider(),
		g.Text("Click the button to run code that produces an error:"),
		g.Button("Trigger Error", func() {
			err := riskyOperation()
			if err != nil {
				errMsg.Set(err.Error())
			}
		}),
		g.When(errMsg.Get() != "", func() g.View {
			return g.VStack(
				g.Text("Error caught:").Bold().Color(g.Red),
				g.Text(errMsg.Get()).Color(g.Gray),
				g.Button("Dismiss", func() { errMsg.Set("") }),
			)
		}),
		g.Spacer(),
	).Padding(16)
})

func riskyOperation() error {
	return fmt.Errorf("simulated failure: database connection refused (port 5432)")
}

var App = g.Define(func(s *g.Scope) g.View {
	g.Provide(s, NotesStore, NotesModel{NextID: 1})

	return g.TabView(
		g.Tab("Notes", "", NotesTab),
		g.Tab("Stats", "", StatsTab),
		g.Tab("Errors", "", ErrorTab),
	).Placement(g.TabBottom)
})

func main() {
	g.RunWithConfig(g.AppConfig{
		Title:  "Notes",
		Width:  500,
		Height: 700,
	}, App)
}
