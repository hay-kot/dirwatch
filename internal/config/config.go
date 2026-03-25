package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"

	"github.com/hay-kot/dirwatch/internal/paths"
)

type Config struct {
	Shell    string         `yaml:"shell"`
	ShellCmd string         `yaml:"shell_cmd"`
	Log      Log            `yaml:"log"`
	Vars     map[string]any `yaml:"vars"`
	Watchers []Watcher      `yaml:"watchers"`
}

type Log struct {
	File   string `yaml:"file"`
	Level  string `yaml:"level"`
	Format string `yaml:"format"`
	Color  bool   `yaml:"color"`
}

type Watcher struct {
	Dirs   []string `yaml:"dirs"`
	Events []string `yaml:"events"`
	Match  []string `yaml:"matches"`
	Exec   string   `yaml:"exec"`
}

func Default() Config {
	return Config{
		Shell:    "/bin/bash",
		ShellCmd: "-c",
		Log: Log{
			Level:  "info",
			Format: "text",
			Color:  false,
		},
		Vars: map[string]any{},
	}
}

// configNames is the list of config file names searched in priority order.
var configNames = []string{
	"dirwatch.yaml",
	"dirwatch.yml",
	"config.yaml",
	"config.yml",
}

// Find locates the config file path by searching for known filenames in the
// config directory. Returns an empty string if no config file is found.
func Find() string {
	dir := paths.ConfigDir()
	for _, name := range configNames {
		p := filepath.Join(dir, name)
		if _, err := os.Stat(p); err == nil {
			return p
		}
	}
	return ""
}

func Read() (Config, error) {
	path := Find()
	if path == "" {
		return Default(), nil
	}
	return ReadFrom(path)
}

func ReadFrom(path string) (Config, error) {
	cfg := Default()

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return cfg, nil
		}
		return cfg, fmt.Errorf("reading config: %w", err)
	}

	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return cfg, fmt.Errorf("parsing config: %w", err)
	}

	dir := filepath.Dir(path)
	for i := range cfg.Watchers {
		for j := range cfg.Watchers[i].Dirs {
			cfg.Watchers[i].Dirs[j] = ExpandPath(dir, cfg.Watchers[i].Dirs[j])
		}
		cfg.Watchers[i].Exec = ExpandPath(dir, cfg.Watchers[i].Exec)
	}

	return cfg, nil
}

func (c Config) Dump() (string, error) {
	b, err := yaml.Marshal(c)
	if err != nil {
		return "", fmt.Errorf("encoding config: %w", err)
	}
	return strings.TrimSpace(string(b)), nil
}
