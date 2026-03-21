package commands

import "github.com/hay-kot/dirwatch/internal/config"

type Flags struct {
	LogLevel   string
	NoColor    bool
	LogFile    string
	ConfigFile string
	Config     config.Config
}

func (f *Flags) LoadConfig() (config.Config, error) {
	if f.ConfigFile != "" {
		return config.ReadFrom(f.ConfigFile)
	}
	return config.Read()
}
