package paths

import (
	"os"
	"path/filepath"
)

const appName = "dirwatch"

func ConfigDir() string {
	if dir := os.Getenv("XDG_CONFIG_HOME"); dir != "" {
		return filepath.Join(dir, appName)
	}
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".config", appName)
}

func DataDir() string {
	if dir := os.Getenv("XDG_DATA_HOME"); dir != "" {
		return filepath.Join(dir, appName)
	}
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".local", "share", appName)
}

func CacheDir() string {
	if dir := os.Getenv("XDG_CACHE_HOME"); dir != "" {
		return filepath.Join(dir, appName)
	}
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".cache", appName)
}
