// Command gova is the developer tool for Gova applications. It wraps the
// Go toolchain with a dev-oriented hot-reload workflow and a convenience
// build/run shortcut.
//
//	gova dev [pkg]    # watch and reload on file change
//	gova build [pkg]  # compile pkg to ./bin/<name>
//	gova run [pkg]    # build and run once
//	gova version
//	gova help
package main

import (
	"context"
	"log"
	"os"

	"github.com/urfave/cli/v3"
)

const version = "0.1.0"

func main() {
	// dev
	// build
	// run
	// version
	// help
	cmd := &cli.Command{
		Name: "gova",
		// TODO: Description
		Description:    "some description",
		Version:        "0.1.0",
		DefaultCommand: "help",
		Commands: []*cli.Command{
			initCommand,
			devCommand,
			buildCmd,
			{
				Name:        "run",
				Description: "Build and run the binary once",
			},
		},
	}

	if err := cmd.Run(context.Background(), os.Args); err != nil {
		log.Fatal(err)
	}
}
