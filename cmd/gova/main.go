package main

import (
	"context"
	"log"
	"os"

	"github.com/urfave/cli/v3"
)

const version = "0.1.0"

func main() {
	cmd := &cli.Command{
		Name:           "gova",
		Description:    "Gova is a GUI toolkit for Go",
		Version:        version,
		DefaultCommand: "help",
		Commands: []*cli.Command{
			initCommand,
			devCommand,
			buildCmd,
			runCmd,
		},
	}

	if err := cmd.Run(context.Background(), os.Args); err != nil {
		log.Fatal(err)
	}
}
