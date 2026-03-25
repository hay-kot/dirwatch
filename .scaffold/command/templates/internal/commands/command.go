package commands

import (
	"context"
	"fmt"

	"github.com/rs/zerolog/log"
	"github.com/urfave/cli/v3"
)

// {{ .Scaffold.command_name | toPascalCase }}Cmd implements the {{ .Scaffold.command_name }} command
type {{ .Scaffold.command_name | toPascalCase }}Cmd struct {
	flags *Flags
}

// New{{ .Scaffold.command_name | toPascalCase }}Cmd creates a new {{ .Scaffold.command_name }} command
func New{{ .Scaffold.command_name | toPascalCase }}Cmd(flags *Flags) *{{ .Scaffold.command_name | toPascalCase }}Cmd {
	return &{{ .Scaffold.command_name | toPascalCase }}Cmd{flags: flags}
}

// Register adds the {{ .Scaffold.command_name }} command to the application
func (cmd *{{ .Scaffold.command_name | toPascalCase }}Cmd) Register(app *cli.Command) *cli.Command {
	app.Commands = append(app.Commands, &cli.Command{
		Name:  "{{ .Scaffold.command_name }}",
		Usage: "{{ .Scaffold.command_name }} command",
		Flags: []cli.Flag{
			// Add command-specific flags here
		},
		Action: cmd.run,
	})

	return app
}

func (cmd *{{ .Scaffold.command_name | toPascalCase }}Cmd) run(ctx context.Context, c *cli.Command) error {
	log.Info().Msg("running {{ .Scaffold.command_name }} command")

	fmt.Println("Hello World!")

	return nil
}
