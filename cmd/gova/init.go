package main

import (
	"context"
	"time"

	"github.com/nv404/gova/internal/config"
	"github.com/urfave/cli/v3"
)

// Init Command initialises a go module to work with the Gova CLI
// It creates a gova.json files with metadata related to building the project and
// running the dev server
var initCommand = &cli.Command{
	Name:        "init",
	Description: "Initialise a gova project",
	Action: func(ctx context.Context, c *cli.Command) error {
		conf := config.Config{
			Version:   "0.0.1",
			Name:      "govaapp",
			Package:   "github.com/cchirag/gova",
			OutputDir: "bin",
			Debounce:  config.Duration{Duration: time.Millisecond * 2000},
			Ignore:    []string{"vendor", "node_modules", ".git"},
			Modes: map[string]config.BuildMode{
				"dev": {
					BuildFlags: []string{},
					EnvFiles:   []string{},
					Env:        []string{},
				},
				"prod": {
					BuildFlags: []string{},
					EnvFiles:   []string{},
					Env:        []string{},
				},
			},
		}
		return config.SetConfig(conf, false)
	},
}
