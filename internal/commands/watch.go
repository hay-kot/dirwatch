package commands

import (
	"context"

	"github.com/urfave/cli/v3"
)

type WatchCmd struct {
	flags *Flags
}

func NewWatchCmd(flags *Flags) *WatchCmd {
	return &WatchCmd{flags: flags}
}

func (cmd *WatchCmd) Register(app *cli.Command) *cli.Command {
	app.Commands = append(app.Commands, &cli.Command{
		Name:   "watch",
		Usage:  "watch directories and run commands on file events",
		Action: cmd.run,
	})
	return app
}

func (cmd *WatchCmd) run(ctx context.Context, c *cli.Command) error {
	return nil
}
