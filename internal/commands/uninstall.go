package commands

import (
	"context"
	"fmt"

	"github.com/urfave/cli/v3"

	"github.com/hay-kot/dirwatch/internal/launchd"
)

type UninstallCmd struct {
	flags *Flags
}

func NewUninstallCmd(flags *Flags) *UninstallCmd {
	return &UninstallCmd{flags: flags}
}

func (cmd *UninstallCmd) Register(app *cli.Command) *cli.Command {
	app.Commands = append(app.Commands, &cli.Command{
		Name:   "uninstall",
		Usage:  "remove dirwatch background service (macOS LaunchAgent)",
		Action: cmd.run,
	})
	return app
}

func (cmd *UninstallCmd) run(ctx context.Context, c *cli.Command) error {
	fmt.Println("Uninstalling dirwatch service...")

	if err := launchd.Uninstall(); err != nil {
		return fmt.Errorf("uninstalling service: %w", err)
	}

	fmt.Println("Service stopped and removed.")
	return nil
}
