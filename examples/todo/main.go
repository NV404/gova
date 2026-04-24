package main

import (
	"fmt"

	g "github.com/nv404/gova"
)

type Todo struct {
	ID   int
	Text string
	Done bool
}

type Model struct {
	Todos  []Todo
	NextID int
}

func addTodo(m Model, text string) Model {
	m.Todos = append(m.Todos, Todo{ID: m.NextID, Text: text})
	m.NextID++
	return m
}

func toggleTodo(m Model, id int, done bool) Model {
	for i := range m.Todos {
		if m.Todos[i].ID == id {
			m.Todos[i].Done = done
		}
	}
	return m
}

func removeTodo(m Model, id int) Model {
	for i := range m.Todos {
		if m.Todos[i].ID == id {
			m.Todos = append(m.Todos[:i], m.Todos[i+1:]...)
			return m
		}
	}
	return m
}

var TodoApp = g.Define(func(s *g.Scope) g.View {
	model := g.State(s, Model{NextID: 1})
	input := g.State(s, "")

	add := func() {
		text := input.Get()
		if text == "" {
			return
		}
		model.Update(func(m Model) Model { return addTodo(m, text) })
		input.Set("")
	}

	todos := g.DerivedList(model, func(m Model) []Todo { return m.Todos })
	itemCount := g.Derived(model, func(m Model) string {
		return fmt.Sprintf("%d items", len(m.Todos))
	})

	return g.Scaffold(
		g.ScrollView(
			g.List(todos,
				func(t Todo) int { return t.ID },
				func(i int, todo Todo) g.View {
					id := todo.ID
					return g.HStack(
						g.Toggle(todo.Done).OnChange(func(done bool) {
							model.Update(func(m Model) Model {
								return toggleTodo(m, id, done)
							})
						}),
						g.Text(todo.Text),
						g.Spacer(),
						g.Button("X", func() {
							model.Update(func(m Model) Model {
								return removeTodo(m, id)
							})
						}).Color(g.Red),
					)
				},
			),
		).Padding(g.SpaceLG),
	).Top(
		g.VStack(
			g.Text("Todos").Font(g.Title).Bold(),
			g.HStack(
				g.TextField(input).
					Placeholder("What needs to be done?").
					OnSubmit(func(val string) { add() }).
					MinHeight(36).
					Grow(),
				g.Button("Add", add),
			).Spacing(g.SpaceSM),
		).Spacing(g.SpaceMD).Padding(g.SpaceLG),
	).Bottom(
		g.Text(itemCount).Font(g.Caption).Color(g.Secondary).Padding(g.SpaceLG),
	)
})

func main() { g.Run("Todo App", TodoApp) }
