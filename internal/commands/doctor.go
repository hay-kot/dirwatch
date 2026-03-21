package commands

import (
	"context"

	"github.com/urfave/cli/v3"
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
	return nil
}
