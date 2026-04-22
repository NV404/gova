package main

import (
	"flag"
	"fmt"
	"io"
	"time"

	"github.com/nv404/gova/internal/devserver"
)

func cmdDev(args []string, stdout, stderr io.Writer) int {
	fs := flag.NewFlagSet("dev", flag.ContinueOnError)
	fs.SetOutput(stderr)
	debounce := fs.Duration("debounce", 200*time.Millisecond, "coalesce window for file-change events")
	stateDir := fs.String("state-dir", "", "directory for persisted dev state (default <workdir>/.gova/dev)")
	pkg, rest, err := parseFlags("dev", fs, args)
	if err != nil {
		return 2
	}
	ctx, cancel := signalContext()
	defer cancel()

	if err := devserver.Run(ctx, devserver.Options{
		Package:  pkg,
		Args:     rest,
		Debounce: *debounce,
		StateDir: *stateDir,
		Stdout:   stdout,
		Stderr:   stderr,
	}); err != nil {
		fmt.Fprintln(stderr, err)
		return 1
	}
	return 0
}
