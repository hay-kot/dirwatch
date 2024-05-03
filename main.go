package main

import (
	"fmt"
	"os"

	"github.com/fsnotify/fsnotify"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/urfave/cli/v2"

	"github.com/hay-kot/watchexec/internal/config"
	"github.com/hay-kot/watchexec/internal/watchhandler"
)

var (
	// Build information. Populated at build-time via -ldflags flag.
	version = "dev"
	commit  = "HEAD"
	date    = "now"
)

func build() string {
	short := commit
	if len(commit) > 7 {
		short = commit[:7]
	}

	return fmt.Sprintf("%s (%s) %s", version, short, date)
}

func main() {
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})

	app := &cli.App{
		Name:    "watchexec",
		Usage:   "Watches a file directory and runs a shell command",
		Version: build(),
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:  "config",
				Usage: "path to the configuration file",
				Value: "watchexec.toml",
			},
		},
		Commands: []*cli.Command{
			{
				Name:  "watch",
				Usage: "watch with configuration file",
				Action: func(ctx *cli.Context) error {
					path := ctx.String("config")

					file, err := os.Open(path)
					if err != nil {
						return err
					}

					cfg, err := config.New(path, file)
					if err != nil {
						return err
					}

					// Create new watcher.
					watcher, err := fsnotify.NewWatcher()
					if err != nil {
						log.Fatal().Err(err).Msg("failed to create watcher")
					}
					defer watcher.Close()

					for _, w := range cfg.Watchers {
						for _, p := range w.Dirs {
							err := watcher.Add(p)
							if err != nil {
								return err
							}
						}
					}

					hdlr := watchhandler.New(cfg)

					// Start listening for events.
					go func() {
						for {
							select {
							case event, ok := <-watcher.Events:
								if !ok {
									return
								}

								if event.Op.Has(fsnotify.Chmod) {
									// Skip Chmod events.
									continue
								}

								log.Debug().
									Str("event", event.Op.String()).
									Str("file_name", event.Name).
									Msg("event")

								hdlr.Handle(event)
							case err, ok := <-watcher.Errors:
								if !ok {
									return
								}
								log.Error().Err(err).Msg("error")
							}
						}
					}()

					// Add a path.
					err = watcher.Add("/tmp")
					if err != nil {
						log.Error().Err(err).Msg("failed to add path")
					}

					// Block main goroutine forever.
					<-make(chan struct{})

					return nil
				},
			},
			{
				Name:   "dev",
				Hidden: true,
				Subcommands: []*cli.Command{
					{
						Name:  "dump",
						Usage: "dump the configuration",
						Action: func(ctx *cli.Context) error {
							path := ctx.String("config")

							file, err := os.Open(path)
							if err != nil {
								return err
							}

							cfg, err := config.New(path, file)
							if err != nil {
								return err
							}

							dump, err := cfg.Dump()
							if err != nil {
								return err
							}

							fmt.Println(dump)
							return nil
						},
					},
				},
			},
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal().Err(err).Msg("failed to run watchexec")
	}
}