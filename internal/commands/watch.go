package commands

import (
	"context"
	"fmt"
	"path/filepath"

	"github.com/fsnotify/fsnotify"
	"github.com/rs/zerolog/log"
	"github.com/urfave/cli/v3"

	"github.com/hay-kot/dirwatch/internal/config"
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
	configPath := cmd.flags.ResolvedConfigPath

	watcher, hdlr, err := setupWatcher(cmd.flags.Config)
	if err != nil {
		return err
	}
	defer func() {
		if err := watcher.Close(); err != nil {
			log.Error().Err(err).Msg("closing watcher")
		}
	}()

	// Watch the config file's directory so we detect atomic writes (new inode).
	if configPath != "" {
		configDir := filepath.Dir(configPath)
		if err := watcher.Add(configDir); err != nil {
			log.Warn().Err(err).Str("path", configDir).Msg("cannot watch config directory for auto-reload")
		} else {
			log.Info().Str("path", configPath).Msg("watching config for changes")
		}
	}

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

			// Check if this is a config file change.
			if configPath != "" && isConfigEvent(event, configPath) {
				if !event.Op.Has(fsnotify.Write) && !event.Op.Has(fsnotify.Create) {
					continue
				}

				log.Info().Msg("config file changed, reloading")
				newCfg, err := config.ReadFrom(configPath)
				if err != nil {
					log.Error().Err(err).Msg("reloading config, keeping current")
					continue
				}

				newWatcher, newHdlr, err := setupWatcher(newCfg)
				if err != nil {
					log.Error().Err(err).Msg("setting up new watchers, keeping current")
					continue
				}

				// Re-watch config directory on the new watcher.
				configDir := filepath.Dir(configPath)
				if err := newWatcher.Add(configDir); err != nil {
					log.Warn().Err(err).Str("path", configDir).Msg("cannot watch config directory on new watcher")
				}

				// Swap: close old watcher, use new one.
				if err := watcher.Close(); err != nil {
					log.Error().Err(err).Msg("closing old watcher")
				}
				watcher = newWatcher
				hdlr = newHdlr
				log.Info().Msg("config reloaded successfully")
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

func setupWatcher(cfg config.Config) (*fsnotify.Watcher, *watchhandler.WatchHandler, error) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, nil, fmt.Errorf("creating watcher: %w", err)
	}

	for _, w := range cfg.Watchers {
		for _, p := range w.Dirs {
			log.Info().Str("path", p).Msg("watching path")
			if err := watcher.Add(p); err != nil {
				_ = watcher.Close()
				return nil, nil, fmt.Errorf("watching %s: %w", p, err)
			}
		}
	}

	return watcher, watchhandler.New(cfg), nil
}

func isConfigEvent(event fsnotify.Event, configPath string) bool {
	// Compare cleaned absolute paths to handle symlinks and trailing slashes.
	eventPath := filepath.Clean(event.Name)
	return eventPath == filepath.Clean(configPath)
}
