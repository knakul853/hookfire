package config

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/BurntSushi/toml"
)

// Config is the parsed config.toml.
type Config struct {
	Targets map[string]Target `toml:"targets"`
}

// Load reads config.toml at path. A missing file is not an error — it returns an
// empty config, since config is optional (flags can supply everything).
func Load(path string) (Config, error) {
	data, err := os.ReadFile(path)
	if errors.Is(err, fs.ErrNotExist) {
		return Config{Targets: map[string]Target{}}, nil
	}
	if err != nil {
		return Config{}, fmt.Errorf("config: read %s: %w", path, err)
	}
	var c Config
	if err := toml.Unmarshal(data, &c); err != nil {
		return Config{}, fmt.Errorf("config: parse %s: %w", path, err)
	}
	if c.Targets == nil {
		c.Targets = map[string]Target{}
	}
	return c, nil
}

// Target returns the target for alias, if present.
func (c Config) Target(alias string) (Target, bool) {
	t, ok := c.Targets[alias]
	return t, ok
}

// DefaultPath returns $XDG_CONFIG_HOME/hookfire/config.toml (or ~/.config/...).
func DefaultPath() (string, error) {
	base := os.Getenv("XDG_CONFIG_HOME")
	if base == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", fmt.Errorf("config: locate home dir: %w", err)
		}
		base = filepath.Join(home, ".config")
	}
	return filepath.Join(base, "hookfire", "config.toml"), nil
}
