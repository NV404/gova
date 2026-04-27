// Package config provides a convenient way to manage creation and parsing of
// config file (gova.json) while handling CLI commands
package config

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"time"

	"github.com/nv404/gova/internal/utils"
)

type Config struct {
	EntrtPoint string        `json:"entry_point"`
	Debounce   time.Duration `json:"debounce"`
	BinDir     string        `json:"bin_dir"`
}

func getConfigFilePath() (string, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return "", err
	}

	return filepath.Join(cwd, "gova.json"), nil
}

func parseConfig(data []byte) (Config, error) {
	config := new(Config)
	if err := json.Unmarshal(data, config); err != nil {
		return Config{}, err
	}
	return *config, nil
}

func GetConfig() (Config, error) {
	path, err := getConfigFilePath()
	if err != nil {
		return Config{}, err
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return Config{}, err
	}

	return parseConfig(data)
}

func SetConfig(config Config, override bool) error {
	path, err := getConfigFilePath()
	if err != nil {
		return err
	}

	if utils.FileExists(path) && override {
		return errors.New("file already exists. cannot override")
	}

	data, err := json.Marshal(config)
	if err != nil {
		return err
	}

	// TODO: Need a create and swap to avoid corruption
	return os.WriteFile(path, data, 0o644)
}
