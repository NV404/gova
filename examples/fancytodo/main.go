package main

import (
	"time"

	g "github.com/nv404/gova"
)

type Category string

const (
	CatAll      Category = "All"
	CatWork     Category = "Work"
	CatPersonal Category = "Personal"
	CatShopping Category = "Shopping"
	CatHealth   Category = "Health"
)

var allCategories = []Category{CatWork, CatPersonal, CatShopping, CatHealth}
var filterChoices = []Category{CatAll, CatWork, CatPersonal, CatShopping, CatHealth}

type Todo struct {
	ID       int
	Text     string
	Category Category
	Done     bool
}

func buildTheme() *g.Theme {
	bg := g.Hex("#0F1116")
	surface := g.Hex("#1B1E26")
	primary := g.Hex("#EDEEF2")
	secondary := g.Hex("#7A8099")
	border := g.Hex("#2A2E38")
	accent := g.Hex("#7C5CFF")

	return g.DarkTheme().
		WithAccent(accent).
		SetColor(g.ColorBackground, bg).
		SetColor(g.ColorSurface, surface).
		SetColor(g.ColorForeground, primary).
		SetColor(g.ColorSecondary, secondary).
		SetColor(g.ColorAccent, accent).
		SetColor(g.ColorBorder, border).
		SetColor(g.ColorInputBackground, surface).
		SetColor(g.ColorInputBorder, border).
		SetColor(g.ColorError, g.Hex("#FF5A7A")).
		SetColor(g.ColorSuccess, g.Hex("#4BD37B"))
}

func categoryDot(c Category) any {
	switch c {
	case CatWork:
		return g.Hex("#5EA9FF")
	case CatPersonal:
		return g.Hex("#FFB454")
	case CatShopping:
		return g.Hex("#FF6AB3")
	case CatHealth:
		return g.Hex("#4BD37B")
	}
	return g.Secondary
}

const fadeMillis = 220

type fadeState struct {
	ID    int     // the item currently animating; 0 = none
	Alpha float64 // 0..1
	Dir   int     // -1 = fading out, +1 = fading in, 0 = idle
}

var FancyTodo = g.Define(func(s *g.Scope) g.View {
	todos := g.State(s, []Todo{
		{ID: 1, Text: "Ship gova v0.2 docs", Category: CatWork, Done: false},
		{ID: 2, Text: "Groceries for the week", Category: CatShopping, Done: false},
		{ID: 3, Text: "30 min walk", Category: CatHealth, Done: true},
	})
	nextID := g.State(s, 4)
	input := g.State(s, "")
	category := g.State(s, CatWork)
	filter := g.State(s, CatAll)

	// Only one row animates at a time; fadeProgress (0..1) drives Opacity on
	// that row. On fade-out completion the row is removed from state.
	fade := g.State(s, fadeState{Alpha: 1})
	errorMsg := g.State(s, "")

	fadeIn := g.UseAnimation(s, g.Animate(fadeMillis*time.Millisecond).WithCurve(g.EaseOut))
	fadeOut := g.UseAnimation(s, g.Animate(fadeMillis*time.Millisecond).WithCurve(g.EaseIn))

	add := func() {
		text := input.Get()
		if text == "" {
			errorMsg.Set("Please enter something first")
			return
		}
		if errorMsg.Get() != "" {
			errorMsg.Set("")
		}
		id := nextID.Get()
		nextID.Set(id + 1)
		g.Append(todos, Todo{ID: id, Text: text, Category: category.Get()})
		input.Set("")

		// Stop any in-flight fade first, then commit the new fade state in one
		// Set so the renderer never sees an intermediate (new id, old alpha).
		fadeIn.Stop()
		fadeOut.Stop()
		fade.Set(fadeState{ID: id, Alpha: 0, Dir: +1})
		fadeIn.Start(func(t float32) {
			cur := fade.Get()
			if cur.ID != id || cur.Dir != +1 {
				return
			}
			fade.Set(fadeState{ID: id, Alpha: float64(t), Dir: +1})
		})
	}

	toggle := func(id int, done bool) {
		g.UpdateWhere(todos,
			func(t Todo) bool { return t.ID == id },
			func(t Todo) Todo { t.Done = done; return t },
		)
	}

	remove := func(id int) {
		cur := fade.Get()
		if cur.ID == id && cur.Dir == -1 {
			return
		}
		fadeIn.Stop()
		fadeOut.Stop()
		fade.Set(fadeState{ID: id, Alpha: 1, Dir: -1})
		fadeOut.Start(func(t float32) {
			cur := fade.Get()
			if cur.ID != id || cur.Dir != -1 {
				return
			}
			if t >= 1 {
				g.RemoveWhere(todos, func(x Todo) bool { return x.ID == id })
				fade.Set(fadeState{Alpha: 1})
				return
			}
			fade.Set(fadeState{ID: id, Alpha: 1 - float64(t), Dir: -1})
		})
	}

	header := g.VStack(
		g.Text("Fancy Todo").Font(g.Title).Bold(),
		g.Text("Keep your week in order").Font(g.Caption).Color(g.Secondary),
	).Spacing(g.SpaceXS).Align(g.Leading)

	addBtn := g.HStack(
		g.Text("+").Color(g.White).Bold(),
		g.Text("Add").Color(g.White).Bold(),
	).Spacing(g.SpaceXS).
		PaddingH(g.SpaceMD).
		PaddingV(g.SpaceXS).
		Background(g.Accent).
		CornerRadius(999).
		OnTap(add)

	composerRow := g.HStack(
		g.PickerOf(category, allCategories, func(c Category) string { return string(c) }).
			MinWidth(110),
		g.TextField(input).
			Placeholder("What needs to be done?").
			OnSubmit(func(string) { add() }).
			Grow(),
		addBtn,
	).Spacing(g.SpaceSM)

	composerChildren := []any{composerRow}
	if errorMsg.Get() != "" {
		composerChildren = append(composerChildren,
			g.Text(errorMsg.Get()).Font(g.Caption).Color(g.Destructive),
		)
	}

	composer := g.VStack(composerChildren...).
		Spacing(g.SpaceXS).
		Padding(g.SpaceMD).
		Background(g.Surface).
		CornerRadius(14).
		Stroke(g.BorderColor, 1)

	chips := g.ScrollView(
		g.HStack(chipRow(filter, filterChoices)...).Spacing(g.SpaceSM),
	).ScrollDirection(g.ScrollHorizontal).PaddingTop(g.SpaceSM)

	top := g.VStack(
		header,
		composer,
		chips,
	).Spacing(g.SpaceMD).
		PaddingH(g.SpaceLG).
		PaddingTop(g.SpaceLG).
		PaddingBottom(g.SpaceSM).
		Background(g.Background)

	current := filter.Get()
	visible := filterTodos(todos.Get(), current)

	curFade := fade.Get()
	var rows []any
	for _, t := range visible {
		t := t
		alpha := 1.0
		if t.ID == curFade.ID && curFade.ID != 0 {
			alpha = curFade.Alpha
		}
		rows = append(rows, todoRow(t, alpha, toggle, remove))
	}

	var listBody g.View
	if len(visible) == 0 {
		listBody = g.VStack(
			g.Spacer(),
			g.Text("Nothing here").Font(g.Title3).Color(g.Secondary),
			g.Text("Pick a different filter or add a new item.").Font(g.Caption).Color(g.Secondary),
			g.Spacer(),
		).Spacing(g.SpaceSM).Align(g.Center).PaddingH(g.SpaceLG).PaddingV(g.SpaceXL)
	} else {
		listBody = g.VStack(rows...).
			Spacing(g.SpaceSM).
			PaddingH(g.SpaceLG).
			PaddingBottom(g.SpaceLG).
			PaddingTop(g.SpaceXS)
	}

	return g.Scaffold(
		g.ScrollView(listBody),
	).Top(top)
})

func filterTodos(xs []Todo, f Category) []Todo {
	if f == CatAll {
		return xs
	}
	out := make([]Todo, 0, len(xs))
	for _, t := range xs {
		if t.Category == f {
			out = append(out, t)
		}
	}
	return out
}

func chipRow(selected *g.StateValue[Category], choices []Category) []any {
	cur := selected.Get()
	out := make([]any, 0, len(choices))
	for _, c := range choices {
		c := c
		active := c == cur
		label := g.Text(string(c)).Font(g.Caption)
		if active {
			label = label.Color(g.White).Bold()
		} else {
			label = label.Color(g.Secondary)
		}
		bg := g.Surface
		stroke := g.BorderColor
		if active {
			bg = g.Accent
			stroke = g.Accent
		}
		chip := g.HStack(label).
			PaddingH(g.SpaceMD).
			PaddingV(g.SpaceSM).
			Background(bg).
			Stroke(stroke, 1).
			CornerRadius(999).
			OnTap(func() { selected.Set(c) })
		out = append(out, chip)
	}
	return out
}

func todoRow(t Todo, alpha float64, onToggle func(int, bool), onDelete func(int)) g.View {
	textColor := g.Primary
	if t.Done {
		textColor = g.Secondary
	}

	title := g.Text(t.Text).
		Color(textColor).
		Strikethrough(t.Done)

	meta := g.HStack(
		g.Text("•").Color(categoryDot(t.Category)).Bold(),
		g.Text(string(t.Category)).Font(g.Caption).Color(g.Secondary),
	).Spacing(g.SpaceXS)

	body := g.VStack(title, meta).Spacing(g.SpaceXS).Align(g.Leading).Grow()

	deleteBtn := g.HStack(
		g.Text("Delete").Font(g.Caption).Color(g.Secondary).Bold(),
	).PaddingH(g.SpaceSM).
		PaddingV(g.SpaceXS).
		Background(g.Hex("#242832")).
		Stroke(g.BorderColor, 1).
		CornerRadius(999).
		OnTap(func() { onDelete(t.ID) })

	return g.HStack(
		g.Toggle(t.Done).OnChange(func(done bool) { onToggle(t.ID, done) }),
		body,
		deleteBtn,
	).Spacing(g.SpaceSM).
		PaddingH(g.SpaceMD).
		PaddingV(g.SpaceSM).
		Background(g.Surface).
		CornerRadius(10).
		Stroke(g.BorderColor, 1).
		Opacity(alpha)
}

func main() {
	g.RunWithConfig(g.AppConfig{
		Title:  "Fancy Todo",
		Width:  480,
		Height: 680,
		Theme:  buildTheme(),
	}, FancyTodo)
}
