package config

import (
	"os"
	"path/filepath"
	"strings"
)

// Config holds user preferences loaded from ~/.config/chop/config.yml.
type Config struct {
	Tee      bool
	Disabled []string
}

// Path returns the config file path, respecting XDG_CONFIG_HOME.
func Path() string {
	dir := os.Getenv("XDG_CONFIG_HOME")
	if dir == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			home = "."
		}
		dir = filepath.Join(home, ".config")
	}
	return filepath.Join(dir, "chop", "config.yml")
}

// Load reads the config file and returns a Config.
// Returns defaults if the file doesn't exist or can't be parsed.
func Load() Config {
	return LoadFrom(Path())
}

// LoadFrom reads config from a specific path. Exported for testing.
func LoadFrom(path string) Config {
	cfg := Config{}

	data, err := os.ReadFile(path)
	if err != nil {
		return cfg
	}

	return parse(string(data))
}

// IsDisabled returns true if the given command is in the disabled list.
func (c Config) IsDisabled(command string) bool {
	for _, d := range c.Disabled {
		if strings.EqualFold(d, command) {
			return true
		}
	}
	return false
}

// parse does simple line-by-line parsing of the config YAML.
func parse(content string) Config {
	cfg := Config{}

	for _, line := range strings.Split(content, "\n") {
		// Strip comments
		if idx := strings.Index(line, "#"); idx >= 0 {
			line = line[:idx]
		}
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		key, value, ok := parseKV(line)
		if !ok {
			continue
		}

		switch key {
		case "tee":
			cfg.Tee = strings.EqualFold(value, "true")
		case "disabled":
			cfg.Disabled = parseList(value)
		}
	}

	return cfg
}

// parseKV splits "key: value" into key and value.
func parseKV(line string) (string, string, bool) {
	idx := strings.Index(line, ":")
	if idx < 0 {
		return "", "", false
	}
	key := strings.TrimSpace(line[:idx])
	value := strings.TrimSpace(line[idx+1:])
	return key, value, true
}

// parseList parses an inline YAML list like "[git, docker]" or "[]".
func parseList(value string) []string {
	value = strings.TrimSpace(value)
	if value == "[]" || value == "" {
		return nil
	}

	// Strip brackets
	value = strings.TrimPrefix(value, "[")
	value = strings.TrimSuffix(value, "]")

	var items []string
	for _, item := range strings.Split(value, ",") {
		item = strings.TrimSpace(item)
		// Strip quotes if present
		item = strings.Trim(item, "\"'")
		if item != "" {
			items = append(items, item)
		}
	}
	return items
}
