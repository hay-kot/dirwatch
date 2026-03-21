package commands

import (
	"context"
	"fmt"

	"github.com/fsnotify/fsnotify"
	"github.com/rs/zerolog/log"
	"github.com/urfave/cli/v3"

	"github.com/hay-kot/dirwatch/internal/watchhandler"
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
	cfg := cmd.flags.Config

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return fmt.Errorf("creating watcher: %w", err)
	}
	defer watcher.Close()

	for _, w := range cfg.Watchers {
		for _, p := range w.Dirs {
			log.Info().Str("path", p).Msg("watching path")
			if err := watcher.Add(p); err != nil {
				return fmt.Errorf("watching %s: %w", p, err)
			}
		}
	}

	hdlr := watchhandler.New(cfg)

	for {
		select {
		case <-ctx.Done():
			log.Info().Msg("shutting down")
			return nil
		case event, ok := <-watcher.Events:
			if !ok {
				return nil
			}
			if event.Op.Has(fsnotify.Chmod) {
				continue
			}
			log.Debug().
				Str("event", event.Op.String()).
				Str("file_name", event.Name).
				Msg("event")
			hdlr.Handle(event)
		case err, ok := <-watcher.Errors:
			if !ok {
				return nil
			}
			log.Error().Err(err).Msg("watcher error")
		}
	}
}
