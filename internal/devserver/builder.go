package devserver

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
)

// Build runs `go build -o out pkg` from the given working directory.
// Returns stderr output on failure so the caller can display it verbatim.
type BuildError struct {
	Stderr string
	Err    error
}

func (e *BuildError) Error() string {
	if e.Stderr == "" {
		return fmt.Sprintf("gova: build failed: %v", e.Err)
	}
	return fmt.Sprintf("gova: build failed: %v\n%s", e.Err, e.Stderr)
}

func (e *BuildError) Unwrap() error { return e.Err }

// Build compiles pkg (a Go package path, e.g. "." or "./cmd/app") to out.
// It uses `go build` in the provided working directory. Context cancellation
// kills the build.
func Build(ctx context.Context, workDir, pkg, out string) error {
	if pkg == "" {
		pkg = "."
	}
	var stderr bytes.Buffer
	cmd := exec.CommandContext(ctx, "go", "build", "-o", out, pkg)
	cmd.Dir = workDir
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return &BuildError{Stderr: stderr.String(), Err: err}
	}
	return nil
}
