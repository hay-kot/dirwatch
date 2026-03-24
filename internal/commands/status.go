package commands

import (
	"context"
	"fmt"

	"github.com/urfave/cli/v3"

	"github.com/hay-kot/dirwatch/internal/launchd"
)

type StatusCmd struct {
	flags *Flags
}

func NewStatusCmd(flags *Flags) *StatusCmd {
	return &StatusCmd{flags: flags}
}

func (cmd *StatusCmd) Register(app *cli.Command) *cli.Command {
	app.Commands = append(app.Commands, &cli.Command{
		Name:   "status",
		Usage:  "show dirwatch background service status",
		Action: cmd.run,
	})
	return app
}

func (cmd *StatusCmd) run(ctx context.Context, c *cli.Command) error {
	s := launchd.GetStatus()

	fmt.Printf("  service: %s\n", launchd.Label)
	fmt.Printf("  state:   %s\n", s.State)
	if s.PID > 0 {
		fmt.Printf("  pid:     %d\n", s.PID)
	}
	fmt.Printf("  plist:   %s\n", launchd.PlistPath())
	fmt.Printf("  binary:  %s\n", launchd.BinPath())

	return nil
}
