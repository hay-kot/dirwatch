package config

import (
	"os"
	"path/filepath"
	"strings"
)

func ExpandPath(confdir, input string) string {
	if strings.HasPrefix(input, "~/") {
		home, err := os.UserHomeDir()
		if err != nil {
			panic(err)
		}
		return filepath.Join(home, input[2:])
	}

	if strings.HasPrefix(input, "./") {
		return filepath.Join(confdir, input[2:])
	}

	return input
}
