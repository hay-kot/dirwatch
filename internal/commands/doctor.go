package commands

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/urfave/cli/v3"

	"github.com/hay-kot/dirwatch/internal/paths"
)

type DoctorCmd struct {
	flags *Flags
}

func NewDoctorCmd(flags *Flags) *DoctorCmd {
	return &DoctorCmd{flags: flags}
}

func (cmd *DoctorCmd) Register(app *cli.Command) *cli.Command {
	app.Commands = append(app.Commands, &cli.Command{
		Name:   "doctor",
		Usage:  "check dirwatch health and configuration",
		Action: cmd.run,
	})
	return app
}

func (cmd *DoctorCmd) run(ctx context.Context, c *cli.Command) error {
	ok := true
	cfg := cmd.flags.Config

	configPath := cmd.flags.ConfigFile
	if configPath == "" {
		configPath = filepath.Join(paths.ConfigDir(), "config.yaml")
	}
	if _, err := os.Stat(configPath); err != nil {
		fmt.Printf("  config: %s (not found)\n", configPath)
		ok = false
	} else {
		fmt.Printf("  config: %s (ok)\n", configPath)
	}

	if _, err := os.Stat(cfg.Shell); err != nil {
		fmt.Printf("  shell: %s (not found)\n", cfg.Shell)
		ok = false
	} else {
		fmt.Printf("  shell: %s (ok)\n", cfg.Shell)
	}

	for _, w := range cfg.Watchers {
		for _, dir := range w.Dirs {
			if info, err := os.Stat(dir); err != nil {
				fmt.Printf("  watch dir: %s (not found)\n", dir)
				ok = false
			} else if !info.IsDir() {
				fmt.Printf("  watch dir: %s (not a directory)\n", dir)
				ok = false
			} else {
				fmt.Printf("  watch dir: %s (ok)\n", dir)
			}
		}
	}

	if !ok {
		return fmt.Errorf("doctor found issues")
	}

	fmt.Println("\nAll checks passed.")
	return nil
}
