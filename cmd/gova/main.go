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
	"flag"
	"fmt"
	"io"
	"os"
	"os/signal"
	"syscall"
)

const version = "0.1.0"

func main() {
	os.Exit(run(os.Args[1:], os.Stdout, os.Stderr))
}

func run(args []string, stdout, stderr io.Writer) int {
	if len(args) == 0 {
		usage(stdout)
		return 2
	}
	switch args[0] {
	case "dev":
		return cmdDev(args[1:], stdout, stderr)
	case "build":
		return cmdBuild(args[1:], stdout, stderr)
	case "run":
		return cmdRun(args[1:], stdout, stderr)
	case "version", "--version", "-v":
		fmt.Fprintf(stdout, "gova %s\n", version)
		return 0
	case "help", "--help", "-h":
		usage(stdout)
		return 0
	default:
		fmt.Fprintf(stderr, "gova: unknown command %q\n\n", args[0])
		usage(stderr)
		return 2
	}
}

func usage(w io.Writer) {
	fmt.Fprint(w, `gova: developer tool for Gova apps

Usage:
  gova <command> [args]

Commands:
  dev [pkg]      Watch the module and rebuild/restart on change (defaults to .)
  build [pkg]    Compile pkg to ./bin/<name>
  run [pkg]      Build and run pkg once
  version        Print the gova CLI version
  help           Show this message

Run "gova <command> --help" for command-specific flags.
`)
}

// signalContext returns a context that cancels on SIGINT/SIGTERM.
func signalContext() (context.Context, context.CancelFunc) {
	ctx, cancel := context.WithCancel(context.Background())
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-ch
		cancel()
		signal.Stop(ch)
	}()
	return ctx, cancel
}

// parseFlags parses the flagset against args, treating the first non-flag arg
// as the package path. Returns the package path (default ".") and leftover
// args (passed to the child).
func parseFlags(name string, fs *flag.FlagSet, args []string) (pkg string, rest []string, err error) {
	fs.Usage = func() {
		fmt.Fprintf(fs.Output(), "usage: gova %s [flags] [pkg] [-- child-args...]\n\n", name)
		fs.PrintDefaults()
	}
	sep := -1
	for i, a := range args {
		if a == "--" {
			sep = i
			break
		}
	}
	head := args
	if sep >= 0 {
		head = args[:sep]
		rest = args[sep+1:]
	}
	if err := fs.Parse(head); err != nil {
		return "", nil, err
	}
	pkg = "."
	if fs.NArg() > 0 {
		pkg = fs.Arg(0)
	}
	return pkg, rest, nil
}
