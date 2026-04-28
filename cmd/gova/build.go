package main

import (
	"context"

	"github.com/nv404/gova/internal/builder"
	"github.com/nv404/gova/internal/config"
	"github.com/urfave/cli/v3"
)

var buildCmd = &cli.Command{
	Name:        "build",
	Description: "Build a gova application",
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:     "mode",
			Required: true,
			Value:    "gova-dev",
			Aliases:  []string{"m"},
			Config: cli.StringConfig{
				TrimSpace: true,
			},
			DefaultText: "-m gova-dev",
			Usage:       "-m dev",
		},
	},
	Action: func(ctx context.Context, c *cli.Command) error {
		mode := c.String("mode")

		cnf, err := config.GetConfig()
		if err != nil {
			return err
		}

		if err := builder.BuildApplication(ctx, mode, cnf); err != nil {
			return err
		}

		return nil
	},
}
