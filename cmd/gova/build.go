package main

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/nv404/gova/internal/config"
	"github.com/nv404/gova/internal/utils"
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

		if err := buildApplication(ctx, mode, cnf); err != nil {
			return err
		}

		return nil
	},
}

// buildApplication compiles the Go package defined in cfg using the build
// configuration for the given mode. It resolves the output binary path,
// sanitizes user-provided build flags, injects environment variables from
// EnvFiles and Env in that order, and invokes `go build`.
//
// The output binary is placed at OutputDir/BinaryName relative to the
// current working directory.
//
// Environment variables are layered as follows:
//   - The current process environment (os.Environ) as the base
//   - EnvFiles merged in order, later files taking precedence
//   - Inline Env values taking final precedence
func buildApplication(ctx context.Context, mode string, cfg config.Config) error {
	m, ok := cfg.Modes[mode]
	if !ok {
		return fmt.Errorf("unknown mode %q", mode)
	}

	wd, err := os.Getwd()
	if err != nil {
		return err
	}

	outputDir := utils.NormalizePath(cfg.OutputDir)
	output := filepath.Join(wd, outputDir, utils.BinaryName(cfg.Name))
	path := strings.Join([]string{".", utils.NormalizePath(cfg.Package)}, string(filepath.Separator))

	flags := append([]string{"build"}, utils.SanitizeBuildFlags(m.BuildFlags)...)
	flags = append(flags, "-o", output, path)

	build := exec.CommandContext(ctx, "go", flags...)

	env, err := utils.ResolveEnv(m)
	if err != nil {
		return err
	}

	build.Env = env
	build.Stdin = os.Stdin
	build.Stdout = os.Stdout
	build.Stderr = os.Stderr

	if err := build.Run(); err != nil {
		return fmt.Errorf("go build failed: %w", err)
	}

	return nil
}
