package watchhandler

import (
	"os/exec"
	"path/filepath"

	"github.com/bmatcuk/doublestar/v4"
	"github.com/fsnotify/fsnotify"
	"github.com/hay-kot/dirwatch/internal/config"
	"github.com/hay-kot/dirwatch/internal/quicktmpl"
	"github.com/rs/zerolog/log"
)

type WatchHandler struct {
	cfg *config.Config
}

func New(cfg *config.Config) *WatchHandler {
	return &WatchHandler{
		cfg: cfg,
	}
}

func (wh *WatchHandler) findWatcher(path string) (config.Watcher, bool) {
	pathdir := filepath.Dir(path)

	for _, w := range wh.cfg.Watchers {
		for _, d := range w.Dirs {
			if d == pathdir {
				return w, true
			}
		}
	}

	return config.Watcher{}, false
}

func EventStrToOp(event fsnotify.Event) string {
	switch {
	case event.Op&fsnotify.Create == fsnotify.Create:
		return "create"
	case event.Op&fsnotify.Write == fsnotify.Write:
		return "write"
	case event.Op&fsnotify.Remove == fsnotify.Remove:
		return "remove"
	case event.Op&fsnotify.Rename == fsnotify.Rename:
		return "rename"
	case event.Op&fsnotify.Chmod == fsnotify.Chmod:
		return "chmod"
	}

	return "unknown"
}

func (wh *WatchHandler) Handle(event fsnotify.Event) {
	w, ok := wh.findWatcher(event.Name)
	if !ok {
		log.Debug().
			Str("event", event.Op.String()).
			Str("file_name", event.Name).
			Msg("no watcher found for event")
		return
	}

	// verify event type
	eventStr := EventStrToOp(event)

	hasEvent := false
	for _, e := range w.Events {
		if e == eventStr {
			hasEvent = true
			break
		}
	}

	if !hasEvent {
		log.Debug().
			Str("event", event.Op.String()).
			Str("file_name", event.Name).
			Msg("no event match found for event")
		return
	}

	// verify path matches
	hasMatch := false
	filename := filepath.Base(event.Name)
	for _, d := range w.Match {
		match, err := doublestar.Match(d, filename)
		if err != nil {
			log.Error().Err(err).Msg("failed to match")
			return
		}

		if match {
			hasMatch = true
			log.Debug().
				Str("file_name", filename).
				Str("pattern", d).
				Msg("match")
			break
		}

		log.Debug().
			Str("file_name", filename).
			Str("pattern", d).
			Msg("no match")
	}

	if !hasMatch {
		log.Debug().
			Str("event", event.Op.String()).
			Str("file_name", event.Name).
			Msg("no file match found for event")
		return
	}

	cmdstr, err := quicktmpl.Render(w.Exec, quicktmpl.Data{
		"Path": event.Name,
		"Vars": wh.cfg.Vars,
	})
	if err != nil {
		return
	}

	cmd := exec.Command(wh.cfg.Shell, wh.cfg.ShellCmd, cmdstr)

	cmd.Stdout = log.Logger
	cmd.Stderr = log.Logger

	err = cmd.Run()
	if err != nil {
		log.Error().Err(err).Msg("failed to run command")
	}
}
