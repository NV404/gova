package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"
	"runtime"

	"github.com/nv404/gova/internal/devserver"
	"github.com/urfave/cli/v3"
)

var buildCmd = &cli.Command{
	Name:        "build",
	Description: "Compile binary to ./bin/<name>",
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:      "output",
			Aliases:   []string{"o"},
			TakesFile: true,
			Config: cli.StringConfig{
				TrimSpace: true,
			},
			DefaultText: path.Join(".", "bin"),
			OnlyOnce:    true,
		},
	},
	Arguments: []cli.Argument{
		&cli.StringArg{
			Name:  "name",
			Value: path.Base("."),
			Config: cli.StringConfig{
				TrimSpace: true,
			},
			UsageText: "todo",
		},
	},
	Action: func(ctx context.Context, c *cli.Command) error {
		output := c.String("output")
		name := c.StringArg("name")

		if err := devserver.Build(ctx, wd, pkg, output); err != nil {
			fmt.Fprintln(stderr, err)
			return 1
		}
	},
}

func cmdBuild(args []string, stdout, stderr io.Writer) int {
	fs := flag.NewFlagSet("build", flag.ContinueOnError)
	fs.SetOutput(stderr)
	out := fs.String("o", "", "output path (default ./bin/<pkg-basename>)")
	pkg, _, err := parseFlags("build", fs, args)
	if err != nil {
		return 2
	}
	wd, err := os.Getwd()
	if err != nil {
		fmt.Fprintln(stderr, err)
		return 1
	}
	output := *out
	if output == "" {
		name := filepath.Base(wd)
		if pkg != "." {
			name = filepath.Base(pkg)
		}
		if runtime.GOOS == "windows" {
			name += ".exe"
		}
		output = filepath.Join(wd, "bin", name)
		if err := os.MkdirAll(filepath.Dir(output), 0o755); err != nil {
			fmt.Fprintln(stderr, err)
			return 1
		}
	}
	ctx, cancel := signalContext()
	defer cancel()
	fmt.Fprintf(stdout, "gova build: wrote %s\n", output)
	return 0
}
