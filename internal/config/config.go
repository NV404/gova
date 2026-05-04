// Package config provides a convenient way to manage creation and parsing of
// config file (gova.json) while handling CLI commands
package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/go-playground/validator/v10"
)

var validate = validator.New()

// Config is the top-level configuration for a Gova project,
// loaded from gova.json in the project root.
type Config struct {
	// Version is the schema version of this config file.
	// Used to handle breaking changes across Gova releases.
	Version string `json:"version" validate:"required"`

	// Name is the binary name used in the build process.
	Name string `json:"name" validate:"required,alphanum"`

	// Package is the Go package path to build, e.g. "./cmd/app".
	// Passed directly to `go build` as the target package.
	Package string `json:"package" validate:"required"`

	// Debounce is the duration to wait after a file change before
	// triggering a rebuild. Accepts human-readable strings like "300ms" or "1s".
	Debounce Duration `json:"debounce"`

	// OutputDir is the root directory where compiled binaries are placed.
	// Each mode outputs to OutputDir/<binary>. Defaults to "./bin".
	OutputDir string `json:"output_dir" validate:"required"`

	// Ignore is a list of paths or glob patterns to exclude from file watching.
	// e.g. ["./vendor", "./testdata", "**/*.pb.go"]
	Ignore []string `json:"ignore"`

	// Modes is a map of named build configurations.
	// The key is the mode name, referenced via `gova build --mode <name>`.
	Modes map[string]BuildMode `json:"modes" validate:"required,gt=0,dive"`
}

// BuildMode defines the build configuration for a specific mode,
// such as "dev", "staging", or "prod".
type BuildMode struct {
	// BuildFlags is a list of flags passed directly to `go build`.
	// e.g. ["-race", "-tags=dev"] or ["-ldflags=-s -w"]
	// Do not include -o; the output path is controlled by Gova via OutputDir.
	BuildFlags []string `json:"build_flags" validate:"dive,startswith=-"`

	// Env is a list of environment variables injected into the build process,
	// in KEY=VALUE format. These are merged with EnvFiles, with Env taking precedence.
	Env []string `json:"env" validate:"dive,contains=="`

	// EnvFiles is a list of .env files to load and inject into the build process.
	// Files are merged in order, with later files taking precedence.
	// e.g. [".env", ".env.local"]
	EnvFiles []string `json:"env_files" validate:"dive,endswith=.env"`
}

// Duration wraps time.Duration to support JSON unmarshaling from
// human-readable strings like "300ms", "1s", or "2m".
// The standard time.Duration type only unmarshals from nanosecond integers.
type Duration struct {
	time.Duration
}

// UnmarshalJSON parses a quoted duration string (e.g. "500ms") into a Duration.
func (d *Duration) UnmarshalJSON(b []byte) error {
	var s string
	if err := json.Unmarshal(b, &s); err != nil {
		return fmt.Errorf("duration must be a string, e.g. \"300ms\": %w", err)
	}
	dur, err := time.ParseDuration(s)
	if err != nil {
		return fmt.Errorf("invalid duration %q: %w", s, err)
	}
	d.Duration = dur
	return nil
}

// MarshalJSON serializes the Duration back to a human-readable string.
func (d Duration) MarshalJSON() ([]byte, error) {
	return json.Marshal(d.String())
}

// getConfigFilePath returns the absolute path to gova.json in the current
// working directory.
func getConfigFilePath() (string, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return "", err
	}
	return filepath.Join(cwd, "gova.json"), nil
}

// parseConfig unmarshals raw JSON bytes into a Config and validates all fields.
// Returns a ValidationErrors if any field constraints are violated.
func parseConfig(data []byte) (Config, error) {
	config := new(Config)
	if err := json.Unmarshal(data, config); err != nil {
		return Config{}, fmt.Errorf("failed to parse gova.json: %w", err)
	}

	// Inject gova-dev mode for default

	config.Modes["gova-dev"] = BuildMode{
		BuildFlags: []string{},
		Env:        []string{"GOVA_DEV=1"},
		EnvFiles:   []string{},
	}

	if err := validate.Struct(config); err != nil {
		return Config{}, formatValidationError(err)
	}
	return *config, nil
}

// GetConfig reads and parses gova.json from the current working directory,
// returning a validated Config or an error if the file is missing or invalid.
func GetConfig() (Config, error) {
	path, err := getConfigFilePath()
	if err != nil {
		return Config{}, err
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return Config{}, fmt.Errorf("could not read gova.json: %w", err)
	}
	return parseConfig(data)
}

// SetConfig writes the given Config to gova.json in the current working directory.
// If override is false and the file already exists, an error is returned.
func SetConfig(config Config, override bool) error {
	path, err := getConfigFilePath()
	if err != nil {
		return err
	}

	if fileExists(path) && !override {
		return errors.New("gova.json already exists, pass override to overwrite")
	}

	if err = validate.Struct(config); err != nil {
		return formatValidationError(err)
	}

	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to serialize config: %w", err)
	}

	// TODO: Need a create and swap to avoid corruption
	return os.WriteFile(path, data, 0o644)
}

// formatValidationError converts validator.ValidationErrors into a human-readable
// error message listing each field and the constraint that was violated.
func formatValidationError(err error) error {
	var ve validator.ValidationErrors
	if !errors.As(err, &ve) {
		return err
	}

	msgs := make([]string, 0, len(ve))
	for _, fe := range ve {
		msgs = append(msgs, fmt.Sprintf("  - %s: failed %q constraint", fe.Field(), fe.Tag()))
	}

	return fmt.Errorf("invalid config:\n%s", strings.Join(msgs, "\n"))
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return !errors.Is(err, os.ErrNotExist)
}
