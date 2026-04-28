package builder

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/nv404/gova/internal/config"
	"github.com/nv404/gova/internal/utils"
)

// BuildApplication compiles the Go package defined in cfg using the build
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
func BuildApplication(ctx context.Context, mode string, cfg config.Config) error {
	m, ok := cfg.Modes[mode]
	if !ok {
		return fmt.Errorf("unknown mode %q", mode)
	}

	cwd, err := os.Getwd()
	if err != nil {
		return err
	}

	outputDir := utils.NormalizePath(cfg.OutputDir)
	output, err := filepath.Rel(cwd, filepath.Join(outputDir, utils.BinaryName(cfg.Name)))
	if err != nil {
		return err
	}

	flags := append([]string{"build"}, utils.SanitizeBuildFlags(m.BuildFlags)...)
	flags = append(flags, "-o", output, utils.NormalizePath(cfg.Package))

	build := exec.CommandContext(ctx, "go", flags...)

	build.Env, err = utils.ResolveEnv(m)
	if err != nil {
		return err
	}

	if err := build.Run(); err != nil {
		return fmt.Errorf("go build failed: %w", err)
	}

	return nil
}
