package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"

	"github.com/nv404/gova/internal/devserver"
)

func cmdRun(args []string, stdout, stderr io.Writer) int {
	fs := flag.NewFlagSet("run", flag.ContinueOnError)
	fs.SetOutput(stderr)
	pkg, rest, err := parseFlags("run", fs, args)
	if err != nil {
		return 2
	}
	wd, err := os.Getwd()
	if err != nil {
		fmt.Fprintln(stderr, err)
		return 1
	}
	binDir, err := os.MkdirTemp("", "gova-run-")
	if err != nil {
		fmt.Fprintln(stderr, err)
		return 1
	}
	defer os.RemoveAll(binDir)
	bin := filepath.Join(binDir, "app")
	if runtime.GOOS == "windows" {
		bin += ".exe"
	}
	ctx, cancel := signalContext()
	defer cancel()
	if err := devserver.Build(ctx, wd, pkg, bin); err != nil {
		fmt.Fprintln(stderr, err)
		return 1
	}
	cmd := exec.CommandContext(ctx, bin, rest...)
	cmd.Stdout = stdout
	cmd.Stderr = stderr
	cmd.Stdin = os.Stdin
	if err := cmd.Run(); err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			return exitErr.ExitCode()
		}
		fmt.Fprintln(stderr, err)
		return 1
	}
	return 0
}
