package main

import (
	"context"
	"fmt"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/nv404/gova/internal/devserver"
	"github.com/urfave/cli/v3"
)

// Watcher watches the directory which has gova.json as the root
// Runner runs the binary from the entry_point using go run

var devCommand = &cli.Command{
	Name:        "dev",
	Description: "Watch the module and rebuild/restart on change (defaults to .)",
	Flags: []cli.Flag{
		&cli.DurationFlag{
			Name:        "debounce",
			Aliases:     []string{"d"},
			DefaultText: "200ms",
			OnlyOnce:    true,
			Value:       time.Millisecond * 200,
			Usage:       "-d 200ms",
			Validator: func(d time.Duration) error {
				if d <= 0 {
					return fmt.Errorf("debounce duration cannot be 0")
				}
				return nil
			},
		},
	},

	Arguments: []cli.Argument{
		&cli.StringArg{
			Name:      "package",
			Value:     ".",
			UsageText: "./cmd/app",
			Config: cli.StringConfig{
				TrimSpace: true,
			},
		},
	},
	Action: func(ctx context.Context, c *cli.Command) error {
		debounce, stateDir, watchDir := c.Duration("debounce"), c.String("state_dir"), c.String("watch-dir")
		pkg := c.StringArg("pkg")
		return devserver.Run(ctx, devserver.Options{
			Package:  pkg,
			Args:     c.Args().Tail(),
			Debounce: debounce,
			StateDir: stateDir,
			WorkDir:  watchDir,
		})
	},
}

type DevServer struct {
	watcher    fsnotify.Watcher
	root       string
	entryPoint string
	ctx        context.Context
}

func NewDevServer(root, entryPoint string) (*DevServer, error) {
	// Check if root and entrypoint are valid paths
	return nil, nil
}
