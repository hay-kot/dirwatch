package commands

import "github.com/hay-kot/dirwatch/internal/config"

type Flags struct {
	LogLevel   string
	NoColor    bool
	LogFile    string
	ConfigFile string
	Config     config.Config

	// ResolvedConfigPath is the absolute path to the config file that was loaded.
	ResolvedConfigPath string
}

func (f *Flags) LoadConfig() (config.Config, error) {
	if f.ConfigFile != "" {
		f.ResolvedConfigPath = f.ConfigFile
		return config.ReadFrom(f.ConfigFile)
	}

	path := config.Find()
	if path == "" {
		return config.Default(), nil
	}

	f.ResolvedConfigPath = path
	return config.ReadFrom(path)
}
