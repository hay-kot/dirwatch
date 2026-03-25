package launchd

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"text/template"

	"github.com/hay-kot/dirwatch/internal/paths"
)

const (
	Label    = "io.dirwatch.agent"
	plistTpl = `<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN"
  "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
  <dict>
    <key>Label</key>
    <string>{{ .Label }}</string>

    <key>ProgramArguments</key>
    <array>
      <string>{{ .BinaryPath }}</string>
{{- if .ConfigPath }}
      <string>--config</string>
      <string>{{ .ConfigPath }}</string>
{{- end }}
      <string>watch</string>
    </array>

    <key>EnvironmentVariables</key>
    <dict>
      <key>HOME</key>
      <string>{{ .Home }}</string>
      <key>PATH</key>
      <string>/usr/local/bin:/usr/bin:/bin:/opt/homebrew/bin</string>
    </dict>

    <key>WorkingDirectory</key>
    <string>{{ .Home }}</string>

    <key>RunAtLoad</key>
    <true/>

    <key>KeepAlive</key>
    <dict>
      <key>SuccessfulExit</key>
      <false/>
    </dict>

    <key>ProcessType</key>
    <string>Background</string>

    <key>LowPriorityIO</key>
    <true/>

    <key>StandardOutPath</key>
    <string>{{ .LogDir }}/dirwatch.log</string>

    <key>StandardErrorPath</key>
    <string>{{ .LogDir }}/dirwatch-errors.log</string>
  </dict>
</plist>
`
)

type plistData struct {
	Label      string
	BinaryPath string
	ConfigPath string
	Home       string
	LogDir     string
}

// BinDir returns the stable binary install directory.
func BinDir() string {
	return filepath.Join(paths.DataDir(), "bin")
}

// BinPath returns the stable binary install path.
func BinPath() string {
	return filepath.Join(BinDir(), "dirwatch")
}

// PlistPath returns the LaunchAgent plist file path.
func PlistPath() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, "Library", "LaunchAgents", Label+".plist")
}

func serviceTarget() string {
	return fmt.Sprintf("gui/%d/%s", os.Getuid(), Label)
}

func domainTarget() string {
	return fmt.Sprintf("gui/%d", os.Getuid())
}

// Install copies the running binary to a stable location, writes the plist,
// and bootstraps the LaunchAgent. Safe to re-run for upgrades.
func Install(configPath string) error {
	if runtime.GOOS != "darwin" {
		return fmt.Errorf("service install is only supported on macOS")
	}

	// Resolve the current binary path through symlinks.
	exe, err := os.Executable()
	if err != nil {
		return fmt.Errorf("finding executable: %w", err)
	}
	exe, err = filepath.EvalSymlinks(exe)
	if err != nil {
		return fmt.Errorf("resolving executable symlink: %w", err)
	}

	// Only bake --config into the plist when explicitly provided.
	// Otherwise the service binary discovers the config via config.Find().
	if configPath != "" {
		configPath, err = filepath.Abs(configPath)
		if err != nil {
			return fmt.Errorf("resolving config path: %w", err)
		}
	}

	// If already running, stop first.
	if IsLoaded() {
		if err := bootout(); err != nil {
			return fmt.Errorf("stopping existing service: %w", err)
		}
	}

	// Copy binary to stable location.
	if err := copyBinary(exe, BinPath()); err != nil {
		return fmt.Errorf("installing binary: %w", err)
	}

	// Generate and write plist.
	plist, err := renderPlist(configPath)
	if err != nil {
		return fmt.Errorf("generating plist: %w", err)
	}

	plistPath := PlistPath()
	if err := os.MkdirAll(filepath.Dir(plistPath), 0o755); err != nil {
		return fmt.Errorf("creating LaunchAgents directory: %w", err)
	}
	if err := os.WriteFile(plistPath, plist, 0o644); err != nil {
		return fmt.Errorf("writing plist: %w", err)
	}

	// Bootstrap the service.
	if err := bootstrap(plistPath); err != nil {
		return fmt.Errorf("bootstrapping service: %w", err)
	}

	return nil
}

// Restart stops and starts the service using launchctl kickstart.
func Restart() error {
	if runtime.GOOS != "darwin" {
		return fmt.Errorf("service restart is only supported on macOS")
	}

	if !IsLoaded() {
		return fmt.Errorf("service is not loaded")
	}

	out, err := exec.Command("launchctl", "kickstart", "-k", serviceTarget()).CombinedOutput()
	if err != nil {
		return fmt.Errorf("%s: %w", strings.TrimSpace(string(out)), err)
	}

	return nil
}

// Uninstall stops the service and removes the plist and installed binary.
func Uninstall() error {
	if runtime.GOOS != "darwin" {
		return fmt.Errorf("service uninstall is only supported on macOS")
	}

	if IsLoaded() {
		if err := bootout(); err != nil {
			return fmt.Errorf("stopping service: %w", err)
		}
	}

	plistPath := PlistPath()
	if err := os.Remove(plistPath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("removing plist: %w", err)
	}

	binPath := BinPath()
	if err := os.Remove(binPath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("removing binary: %w", err)
	}

	return nil
}

// Status holds information about the running service.
type Status struct {
	Loaded bool
	PID    int    // 0 if not running
	State  string // e.g. "running", "waiting", "not loaded"
}

// GetStatus returns the current service status.
func GetStatus() Status {
	if runtime.GOOS != "darwin" {
		return Status{State: "unsupported platform"}
	}

	if !IsLoaded() {
		return Status{State: "not loaded"}
	}

	out, err := exec.Command("launchctl", "print", serviceTarget()).CombinedOutput()
	if err != nil {
		return Status{Loaded: true, State: "unknown"}
	}

	s := Status{Loaded: true, State: "loaded"}
	for line := range strings.SplitSeq(string(out), "\n") {
		line = strings.TrimSpace(line)
		if v, ok := strings.CutPrefix(line, "pid = "); ok {
			pid, _ := strconv.Atoi(v)
			s.PID = pid
		}
		if v, ok := strings.CutPrefix(line, "state = "); ok {
			s.State = v
		}
	}

	return s
}

// IsLoaded returns true if the service is currently loaded in launchd.
func IsLoaded() bool {
	err := exec.Command("launchctl", "print", serviceTarget()).Run()
	return err == nil
}

func bootstrap(plistPath string) error {
	out, err := exec.Command("launchctl", "bootstrap", domainTarget(), plistPath).CombinedOutput()
	if err != nil {
		return fmt.Errorf("%s: %w", strings.TrimSpace(string(out)), err)
	}
	return nil
}

func bootout() error {
	out, err := exec.Command("launchctl", "bootout", serviceTarget()).CombinedOutput()
	if err != nil {
		return fmt.Errorf("%s: %w", strings.TrimSpace(string(out)), err)
	}
	return nil
}

func renderPlist(configPath string) ([]byte, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("getting home directory: %w", err)
	}

	logDir := filepath.Join(home, "Library", "Logs")

	data := plistData{
		Label:      Label,
		BinaryPath: BinPath(),
		ConfigPath: configPath,
		Home:       home,
		LogDir:     logDir,
	}

	tmpl, err := template.New("plist").Parse(plistTpl)
	if err != nil {
		return nil, fmt.Errorf("parsing plist template: %w", err)
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return nil, fmt.Errorf("rendering plist: %w", err)
	}

	return buf.Bytes(), nil
}

func copyBinary(src, dst string) error {
	if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
		return fmt.Errorf("creating bin directory: %w", err)
	}

	in, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("opening source binary: %w", err)
	}
	defer func() { _ = in.Close() }()

	// Write to temp file then rename for atomicity.
	tmp := dst + ".tmp"
	out, err := os.OpenFile(tmp, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o755)
	if err != nil {
		return fmt.Errorf("creating temp binary: %w", err)
	}

	cleanup := func() { _ = os.Remove(tmp) }

	if _, err := io.Copy(out, in); err != nil {
		_ = out.Close()
		cleanup()
		return fmt.Errorf("copying binary: %w", err)
	}

	if err := out.Close(); err != nil {
		cleanup()
		return fmt.Errorf("closing temp binary: %w", err)
	}

	if err := os.Rename(tmp, dst); err != nil {
		cleanup()
		return fmt.Errorf("moving binary into place: %w", err)
	}

	return nil
}
