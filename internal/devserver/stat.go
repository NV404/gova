package devserver

import "os"

// statDir is a thin shim so tests can stub filesystem inspection if needed.
// Kept in its own file to keep watcher.go focused on logic.
func statDir(path string) (os.FileInfo, error) { return os.Stat(path) }
