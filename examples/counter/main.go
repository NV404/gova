package main

import g "github.com/nv404/gova"

var Counter = g.Define(func(s *g.Scope) g.View {
	count := g.State(s, 0)
	return g.VStack(
		g.Text(count.Format("Count: %d")),
		g.Button("+", func() { count.Update(func(n int) int { return n + 1 }) }),
		g.Button("-", func() { count.Update(func(n int) int { return n - 1 }) }),
	)
})

func main() { g.Run("Counter", Counter) }
