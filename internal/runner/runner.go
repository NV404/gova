package runner

import (
	"context"
	"fmt"
	"os/exec"
	"path/filepath"

	"github.com/nv404/gova/internal/config"
	"github.com/nv404/gova/internal/utils"
)

// RunApplication executes the compiled binary for the given mode, injecting
// the mode's environment variables. The binary is expected to exist at
// OutputDir/BinaryName. The process is started and returned immediately
// without waiting for it to exit.
//
// args are passed directly to the binary at runtime, e.g. ["--port", "8080"].
func RunApplication(ctx context.Context, mode string, cfg config.Config, args ...string) (*exec.Cmd, error) {
	m, ok := cfg.Modes[mode]
	if !ok {
		return nil, fmt.Errorf("unknown mode %q", mode)
	}

	bin := utils.NormalizePath(filepath.Join(cfg.OutputDir, utils.BinaryName(cfg.Name)))

	env, err := utils.ResolveEnv(m)
	if err != nil {
		return nil, err
	}

	run := exec.CommandContext(ctx, bin, args...)
	run.Env = env

	if err := run.Start(); err != nil {
		return nil, fmt.Errorf("failed to start %q: %w", bin, err)
	}

	return run, nil
}
