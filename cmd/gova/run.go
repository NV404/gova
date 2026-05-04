package main

import (
	"context"
	"fmt"
	"os/exec"
	"path/filepath"

	"github.com/nv404/gova/internal/config"
	"github.com/nv404/gova/internal/utils"
	"github.com/urfave/cli/v3"
)

var runCmd = &cli.Command{
	Name:        "run",
	Description: "Build and run a gova application",
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

		if err := buildApplication(ctx, mode, cnf); err != nil {
			return err
		}

		if _, err := runApplication(ctx, cnf, c.Args().Slice()...); err != nil {
			return err
		}

		return nil
	},
}

// runApplication executes the compiled binary for the given mode, injecting
// the mode's environment variables. The binary is expected to exist at
// OutputDir/BinaryName. The process is started and returned immediately
// without waiting for it to exit.
//
// args are passed directly to the binary at runtime, e.g. ["--port", "8080"].
func runApplication(ctx context.Context, cfg config.Config, args ...string) (*exec.Cmd, error) {
	bin := utils.NormalizePath(filepath.Join(cfg.OutputDir, utils.BinaryName(cfg.Name)))

	run := exec.CommandContext(ctx, bin, args...)

	if err := run.Start(); err != nil {
		return nil, fmt.Errorf("failed to start %q: %w", bin, err)
	}

	return run, nil
}
