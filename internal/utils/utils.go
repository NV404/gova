// Package utils provides shared helper utilities used across Gova's internal
// packages. It covers file system checks, debouncing, environment variable
// resolution, path normalization, and OS-aware binary naming.
package utils

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/hashicorp/go-envparse"
	"github.com/nv404/gova/internal/config"
)

// FileExists reports whether a file or directory exists at the given path.
func FileExists(path string) bool {
	_, err := os.Stat(path)
	return !errors.Is(err, os.ErrNotExist)
}

// Debounce returns a debounced version of fn that delays execution by the
// given duration. If the returned function is called again before the delay
// elapses, the timer resets. Only the last call within a burst will invoke fn.
//
// The returned function is safe to call from multiple goroutines.
func Debounce(fn func(), delay time.Duration) func() {
	var timer *time.Timer
	ch := make(chan struct{}, 1)

	go func() {
		for range ch {
			if timer != nil {
				timer.Stop()
			}
			timer = time.AfterFunc(delay, fn)
		}
	}()

	return func() {
		select {
		case ch <- struct{}{}:
		default:
		}
	}
}

// ResolveEnv builds the final environment variable slice for a given BuildMode.
// It starts with the current process environment as a base, then layers
// EnvFiles in order, and finally applies inline Env values which take
// the highest precedence.
//
// The resulting slice is in "KEY=VALUE" format, compatible with exec.Cmd.Env.
func ResolveEnv(m config.BuildMode) ([]string, error) {
	env := os.Environ()

	for _, file := range m.EnvFiles {
		envs, err := ParseEnvFile(NormalizePath(file))
		if err != nil {
			return nil, fmt.Errorf("failed to parse env file %q: %w", file, err)
		}
		env = append(env, envs...)
	}

	env = append(env, m.Env...)

	return env, nil
}

// ParseEnvFile reads and parses a .env file at the given path, returning
// the key-value pairs as a slice of "KEY=VALUE" strings compatible with
// exec.Cmd.Env. Parsing is handled by go-envparse, which supports quoted
// values, comments, and blank lines.
func ParseEnvFile(name string) ([]string, error) {
	file, err := os.Open(name)
	if err != nil {
		return nil, fmt.Errorf("could not open env file: %w", err)
	}
	defer file.Close()

	parsed, err := envparse.Parse(file)
	if err != nil {
		return nil, fmt.Errorf("could not parse env file: %w", err)
	}

	envs := make([]string, 0, len(parsed))
	for key, value := range parsed {
		envs = append(envs, key+"="+value)
	}

	return envs, nil
}

// SanitizeBuildFlags removes any -o flag from user-provided build flags.
// This prevents the user from overriding the output path that Gova computes
// from OutputDir and BinaryName. Both forms are handled:
//   - "-o", "./path"  (two separate arguments)
//   - "-o=./path"     (single argument with equals sign)
func SanitizeBuildFlags(flags []string) []string {
	result := make([]string, 0, len(flags))
	for i := 0; i < len(flags); i++ {
		if flags[i] == "-o" {
			i++
			continue
		}
		if strings.HasPrefix(flags[i], "-o=") {
			continue
		}
		result = append(result, flags[i])
	}
	return result
}

// NormalizePath converts any user-provided path to the current OS path format,
// handling both forward slashes (POSIX) and backslashes (Windows) as input.
// filepath.Clean is also applied to resolve any redundant separators or dots.
//
// This ensures paths from gova.json work correctly regardless of which OS
// authored the config file.
func NormalizePath(path string) string {
	return filepath.Clean(filepath.FromSlash(path))
}

// BinaryName appends .exe to the binary name when running on Windows.
// On all other platforms the name is returned unchanged.
func BinaryName(name string) string {
	if runtime.GOOS == "windows" {
		return name + ".exe"
	}
	return name
}
