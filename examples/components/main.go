package main

import (
	"fmt"

	g "github.com/nv404/gova"
)

type User struct {
	ID    int
	Name  string
	Email string
}

type UserRow struct {
	User  User
	OnTap func()
}

func (v UserRow) Body(s *g.Scope) g.View {
	return g.HStack(
		g.VStack(
			g.Text(v.User.Name).Bold(),
			g.Text(v.User.Email).Color(g.Secondary).Font(g.Caption),
		),
		g.Spacer(),
		g.Text(">").Color(g.Secondary),
	).Padding(g.SpaceMD).OnTap(v.OnTap)
}

// Badge overlays a small red dot on top of a child view using ZStack.
type Badge struct {
	Child g.View
	Count int
}

func (b Badge) Body(s *g.Scope) g.View {
	return g.ZStack(
		b.Child,
		g.Text(fmt.Sprintf("%d", b.Count)).
			Color(g.White).
			Background(g.Destructive).
			Padding(g.SpaceXS).
			CornerRadius(8),
	).Align(g.TopTrailing)
}

// DetailView is pushed onto the nav stack; captured by value at push time.
type DetailView struct {
	User User
}

func (d DetailView) Body(s *g.Scope) g.View {
	nav := g.UseNav(s)
	return g.VStack(
		g.Text(d.User.Name).Font(g.Title).Bold(),
		g.Text(d.User.Email).Color(g.Secondary),
		g.Spacer(),
		g.Button("Back", nav.Pop),
	).Padding(g.SpaceLG).NavTitle(d.User.Name)
}

// HomeView is the NavStack root.
type HomeView struct{}

func (HomeView) Body(s *g.Scope) g.View {
	users := g.State(s, []User{
		{ID: 1, Name: "Alice", Email: "alice@example.com"},
		{ID: 2, Name: "Bob", Email: "bob@example.com"},
		{ID: 3, Name: "Carol", Email: "carol@example.com"},
	})

	nav := g.UseNav(s)
	addUser := func() {
		next := users.Get()
		id := len(next) + 1
		g.Append(users, User{
			ID:    id,
			Name:  fmt.Sprintf("User %d", id),
			Email: fmt.Sprintf("user%d@example.com", id),
		})
	}

	return g.ScrollView(
		g.ForEach(users.Get(), func(u User) any {
			return UserRow{User: u, OnTap: func() { nav.Push(DetailView{User: u}) }}
		}),
	).NavTitle("Users").NavToolbar(
		g.ToolbarTrailing(g.Button("+", addUser)),
	)
}

var App = g.Define(func(s *g.Scope) g.View {
	return g.NavStack(HomeView{})
})

func main() {
	g.RunWithConfig(g.AppConfig{
		Title:  "Components Demo",
		Width:  380,
		Height: 600,
	}, App)
}
