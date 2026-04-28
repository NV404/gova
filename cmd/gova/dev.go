package main

import (
	"context"
	"os"

	"github.com/nv404/gova/internal/config"
	"github.com/nv404/gova/internal/dev"
	"github.com/urfave/cli/v3"
)

var devCommand = &cli.Command{
	Name:        "dev",
	Description: "Watch the module and rebuild/restart on change (defaults to .)",
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:    "mode",
			Value:   "gova-dev",
			Aliases: []string{"m"},
			Config: cli.StringConfig{
				TrimSpace: true,
			},
			DefaultText: "-m gova-dev",
			Usage:       "-m dev",
		},
	},
	Action: func(ctx context.Context, c *cli.Command) error {
		mode := c.String("mode")
		cfg, err := config.GetConfig()
		if err != nil {
			return err
		}

		cwd, err := os.Getwd()
		if err != nil {
			return err
		}

		server, err := dev.NewDevServer(ctx, cwd, cfg.Package, cfg.Ignore, cfg.Debounce.Duration)
		if err != nil {
			return err
		}

		server.Start(mode, cfg, c.Args().Slice()...)
		<-ctx.Done()
		return nil
	},
}
