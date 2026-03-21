package paths

import (
	"os"
	"path/filepath"
	"testing"
)

func TestConfigDir(t *testing.T) {
	t.Run("XDG override", func(t *testing.T) {
		t.Setenv("XDG_CONFIG_HOME", "/tmp/xdg-config")
		got := ConfigDir()
		want := "/tmp/xdg-config/dirwatch"
		if got != want {
			t.Errorf("ConfigDir() = %q, want %q", got, want)
		}
	})

	t.Run("fallback", func(t *testing.T) {
		t.Setenv("XDG_CONFIG_HOME", "")
		home, _ := os.UserHomeDir()
		got := ConfigDir()
		want := filepath.Join(home, ".config", "dirwatch")
		if got != want {
			t.Errorf("ConfigDir() = %q, want %q", got, want)
		}
	})
}

func TestDataDir(t *testing.T) {
	t.Run("XDG override", func(t *testing.T) {
		t.Setenv("XDG_DATA_HOME", "/tmp/xdg-data")
		got := DataDir()
		want := "/tmp/xdg-data/dirwatch"
		if got != want {
			t.Errorf("DataDir() = %q, want %q", got, want)
		}
	})

	t.Run("fallback", func(t *testing.T) {
		t.Setenv("XDG_DATA_HOME", "")
		home, _ := os.UserHomeDir()
		got := DataDir()
		want := filepath.Join(home, ".local", "share", "dirwatch")
		if got != want {
			t.Errorf("DataDir() = %q, want %q", got, want)
		}
	})
}

func TestCacheDir(t *testing.T) {
	t.Run("XDG override", func(t *testing.T) {
		t.Setenv("XDG_CACHE_HOME", "/tmp/xdg-cache")
		got := CacheDir()
		want := "/tmp/xdg-cache/dirwatch"
		if got != want {
			t.Errorf("CacheDir() = %q, want %q", got, want)
		}
	})

	t.Run("fallback", func(t *testing.T) {
		t.Setenv("XDG_CACHE_HOME", "")
		home, _ := os.UserHomeDir()
		got := CacheDir()
		want := filepath.Join(home, ".cache", "dirwatch")
		if got != want {
			t.Errorf("CacheDir() = %q, want %q", got, want)
		}
	})
}
