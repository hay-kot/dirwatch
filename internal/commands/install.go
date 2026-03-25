package commands

import (
	"context"
	"fmt"

	"github.com/urfave/cli/v3"

	"github.com/hay-kot/dirwatch/internal/launchd"
)

type InstallCmd struct {
	flags *Flags
}

func NewInstallCmd(flags *Flags) *InstallCmd {
	return &InstallCmd{flags: flags}
}

func (cmd *InstallCmd) Register(app *cli.Command) *cli.Command {
	app.Commands = append(app.Commands, &cli.Command{
		Name:  "install",
		Usage: "install dirwatch as a background service (macOS LaunchAgent)",
		Description: `Copies the current binary to a stable location and registers a LaunchAgent.
Safe to re-run after upgrading to update the installed binary.

By default the service discovers the config file automatically. To pin a
specific config file, use the global --config flag:

  dirwatch --config ~/.config/dirwatch/custom.yaml install`,
		Action: cmd.run,
	})
	return app
}

func (cmd *InstallCmd) run(ctx context.Context, c *cli.Command) error {
	fmt.Println("Installing dirwatch service...")

	if err := launchd.Install(cmd.flags.ConfigFile); err != nil {
		return fmt.Errorf("installing service: %w", err)
	}

	fmt.Printf("  binary:  %s\n", launchd.BinPath())
	fmt.Printf("  plist:   %s\n", launchd.PlistPath())
	fmt.Printf("  label:   %s\n", launchd.Label)
	if cmd.flags.ConfigFile != "" {
		fmt.Printf("  config:  %s (pinned)\n", cmd.flags.ConfigFile)
	} else {
		fmt.Println("  config:  auto-discovered")
	}
	fmt.Println("\nService installed and running.")
	fmt.Println("Re-run this command after upgrading to update the installed binary.")
	return nil
}
