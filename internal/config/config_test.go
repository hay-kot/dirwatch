package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestReadFrom_FileNotFound(t *testing.T) {
	cfg, err := ReadFrom("/nonexistent/path/config.yaml")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	def := Default()
	if cfg.Shell != def.Shell {
		t.Errorf("Shell = %q, want %q", cfg.Shell, def.Shell)
	}
	if cfg.Log.Level != def.Log.Level {
		t.Errorf("Log.Level = %q, want %q", cfg.Log.Level, def.Log.Level)
	}
}

func TestReadFrom_ValidYAML(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.yaml")

	content := `
shell: /bin/zsh
shell_cmd: "-c"
log:
  file: /tmp/dirwatch.log
  level: debug
  format: json
  color: true
vars:
  myvar: hello
watchers:
  - dirs:
      - /tmp/watch
    events:
      - create
    matches:
      - "*.txt"
    exec: echo hello
`
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}

	cfg, err := ReadFrom(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if cfg.Shell != "/bin/zsh" {
		t.Errorf("Shell = %q, want /bin/zsh", cfg.Shell)
	}
	if cfg.Log.File != "/tmp/dirwatch.log" {
		t.Errorf("Log.File = %q, want /tmp/dirwatch.log", cfg.Log.File)
	}
	if cfg.Log.Level != "debug" {
		t.Errorf("Log.Level = %q, want debug", cfg.Log.Level)
	}
	if cfg.Log.Format != "json" {
		t.Errorf("Log.Format = %q, want json", cfg.Log.Format)
	}
	if !cfg.Log.Color {
		t.Error("Log.Color = false, want true")
	}
	if len(cfg.Watchers) != 1 {
		t.Fatalf("len(Watchers) = %d, want 1", len(cfg.Watchers))
	}
	if cfg.Watchers[0].Exec != "echo hello" {
		t.Errorf("Watchers[0].Exec = %q, want \"echo hello\"", cfg.Watchers[0].Exec)
	}
	if cfg.Vars["myvar"] != "hello" {
		t.Errorf("Vars[myvar] = %v, want hello", cfg.Vars["myvar"])
	}
}

func TestReadFrom_InvalidYAML(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.yaml")

	if err := os.WriteFile(path, []byte("{{invalid yaml"), 0o644); err != nil {
		t.Fatal(err)
	}

	_, err := ReadFrom(path)
	if err == nil {
		t.Fatal("expected error for invalid YAML, got nil")
	}
}

func TestDump_RoundTrip(t *testing.T) {
	original := Default()
	original.Watchers = []Watcher{
		{
			Dirs:   []string{"/tmp/a"},
			Events: []string{"create"},
			Match:  []string{"*.go"},
			Exec:   "go build",
		},
	}

	dumped, err := original.Dump()
	if err != nil {
		t.Fatalf("Dump() error: %v", err)
	}

	// Write dumped YAML to a temp file and read it back
	dir := t.TempDir()
	path := filepath.Join(dir, "config.yaml")
	if err := os.WriteFile(path, []byte(dumped), 0o644); err != nil {
		t.Fatal(err)
	}

	loaded, err := ReadFrom(path)
	if err != nil {
		t.Fatalf("ReadFrom() error: %v", err)
	}

	if loaded.Shell != original.Shell {
		t.Errorf("Shell = %q, want %q", loaded.Shell, original.Shell)
	}
	if loaded.Log.Level != original.Log.Level {
		t.Errorf("Log.Level = %q, want %q", loaded.Log.Level, original.Log.Level)
	}
	if len(loaded.Watchers) != 1 {
		t.Fatalf("len(Watchers) = %d, want 1", len(loaded.Watchers))
	}
	if loaded.Watchers[0].Exec != original.Watchers[0].Exec {
		t.Errorf("Watchers[0].Exec = %q, want %q", loaded.Watchers[0].Exec, original.Watchers[0].Exec)
	}
}

func TestDefault(t *testing.T) {
	cfg := Default()

	if cfg.Shell != "/bin/bash" {
		t.Errorf("Shell = %q, want /bin/bash", cfg.Shell)
	}
	if cfg.ShellCmd != "-c" {
		t.Errorf("ShellCmd = %q, want -c", cfg.ShellCmd)
	}
	if cfg.Log.Level != "info" {
		t.Errorf("Log.Level = %q, want info", cfg.Log.Level)
	}
	if cfg.Log.Format != "text" {
		t.Errorf("Log.Format = %q, want text", cfg.Log.Format)
	}
	if cfg.Log.Color {
		t.Error("Log.Color = true, want false")
	}
	if cfg.Vars == nil {
		t.Error("Vars is nil, want initialized map")
	}
}
