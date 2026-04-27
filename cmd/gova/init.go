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
			EntrtPoint: "",
			Debounce:   200 * time.Millisecond,
		}
		return config.SetConfig(conf, false)
	},
}
