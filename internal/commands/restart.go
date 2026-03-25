package commands

import (
	"context"
	"fmt"

	"github.com/urfave/cli/v3"

	"github.com/hay-kot/dirwatch/internal/launchd"
)

type RestartCmd struct {
	flags *Flags
}

func NewRestartCmd(flags *Flags) *RestartCmd {
	return &RestartCmd{flags: flags}
}

func (cmd *RestartCmd) Register(app *cli.Command) *cli.Command {
	app.Commands = append(app.Commands, &cli.Command{
		Name:   "restart",
		Usage:  "restart the dirwatch background service",
		Action: cmd.run,
	})
	return app
}

func (cmd *RestartCmd) run(ctx context.Context, c *cli.Command) error {
	if err := launchd.Restart(); err != nil {
		return fmt.Errorf("restarting service: %w", err)
	}

	fmt.Println("Service restarted.")
	return nil
}
