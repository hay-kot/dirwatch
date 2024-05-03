package config

import (
	"io"
	"strings"

	"github.com/BurntSushi/toml"
	"github.com/rs/zerolog"
)

type Config struct {
	Shell    string `toml:"shell"`
	ShellCmd string `toml:"shell_cmd"`

	Log      Log            `toml:"log"`
	Vars     map[string]any `toml:"vars"`
	Watchers []Watcher      `toml:"watchers"`
}

func Default() *Config {
	return &Config{
		Shell:    "/bin/bash",
		ShellCmd: "-c",
		Log: Log{
			Level:  zerolog.InfoLevel,
			Format: "text",
			Color:  false,
		},
		Vars: map[string]any{},
	}
}

func New(confpath string, reader io.Reader) (*Config, error) {
	cfg := Default()
	_, err := toml.NewDecoder(reader).Decode(cfg)
	if err != nil {
		return nil, err
	}

	// Expand paths
	for i := range cfg.Watchers {
		for j := range cfg.Watchers[i].Dirs {
			cfg.Watchers[i].Dirs[j] = ExpandPath(confpath, cfg.Watchers[i].Dirs[j])
		}

		cfg.Watchers[i].Exec = ExpandPath(confpath, cfg.Watchers[i].Exec)
	}

	return cfg, nil
}

// Dump returns the configuration as a TOML string.
func (c Config) Dump() (string, error) {
	var b strings.Builder
	enc := toml.NewEncoder(&b)
	err := enc.Encode(c)
	return b.String(), err
}

type Log struct {
	File   string        `toml:"file"`
	Level  zerolog.Level `toml:"level"`
	Format string        `toml:"format"`
	Color  bool          `toml:"color"`
}

type Watcher struct {
	Dirs   []string `toml:"dirs"`
	Events []string `toml:"events"`
	Match  []string `toml:"matches"`
	Exec   string   `toml:"exec"`
}
