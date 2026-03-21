package commands

import (
	"context"

	"github.com/urfave/cli/v3"
)

type ConfigCmd struct {
	flags *Flags
}

func NewConfigCmd(flags *Flags) *ConfigCmd {
	return &ConfigCmd{flags: flags}
}

func (cmd *ConfigCmd) Register(app *cli.Command) *cli.Command {
	app.Commands = append(app.Commands, &cli.Command{
		Name:   "config",
		Usage:  "show the current configuration",
		Action: cmd.run,
	})
	return app
}

func (cmd *ConfigCmd) run(ctx context.Context, c *cli.Command) error {
	return nil
}
