# Contributing to Gova

Thanks for wanting to help. Gova is pre-1.0, so the bar for
contributions is simple: **the tests pass, the example apps still run,
and the change has a clear motivation.**

## Setup

```bash
git clone https://github.com/nv404/gova.git
cd gova
go test ./...
```

You need:

- **Go 1.26+** (the version declared in `go.mod`).
- **A C toolchain** — cgo is required for the macOS dock / dialog
  bridges. `xcode-select --install` on macOS is enough.
- **Node 20+** — only if you want to work on the docs site at
  `docs-site/`.

## Running the examples

```bash
go run ./examples/counter
go run ./examples/dialogs
```

Every example is self-contained. If an example stops working on
`main`, that is a release blocker — please report it.

## What to work on

Good first contributions:

- Add a widget modifier (colors, padding, shadow — see `modifier.go`).
- Improve platform coverage (Windows / Linux dock + native dialogs).
- Add examples of real apps — a note editor, a log viewer, a timer.
- Improve docs. Every unclear sentence is a bug.

Before a large change, open an issue describing the problem. We try to
avoid design work in review.

## Pull requests

- Keep PRs focused. One change per PR.
- Run `go test ./...` locally. CI runs the same matrix.
- Add a test for behavior change. `harness_test.go` has helpers for
  rendering views without starting a window.
- Update docs under `docs-site/content/docs/` when public API changes.
- Commit messages: imperative mood, ≤70 char summary, body if needed.

## Code style

- Prefer plain Go. Generics are fine; reflection is fine when it
  clearly wins DX (see `state_slice.go`). Avoid unnecessary
  abstraction.
- Comments explain *why*, not *what*. Exported symbols get a doc
  comment; internals usually don't.
- Errors at boundaries (file I/O, external APIs). Inside the package,
  trust your own code.

## Releases

We tag `vX.Y.Z` against `main`. Breaking changes bump `Y` until
`v1.0.0`; after that they need a major.

## License

By contributing, you agree that your code is licensed under the
project's [MIT license](LICENSE).
