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
	cfg config.Config
}

func New(cfg config.Config) *WatchHandler {
	return &WatchHandler{
		cfg: cfg,
	}
}

func (wh *WatchHandler) findWatchers(path string) []config.Watcher {
	pathdir := filepath.Dir(path)

	var watchers []config.Watcher
	for _, w := range wh.cfg.Watchers {
		for _, d := range w.Dirs {
			if d == pathdir {
				watchers = append(watchers, w)
				break
			}
		}
	}

	return watchers
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
	watchers := wh.findWatchers(event.Name)
	if len(watchers) == 0 {
		log.Debug().
			Str("event", event.Op.String()).
			Str("file_name", event.Name).
			Msg("no watcher found for event")
		return
	}

	eventStr := EventStrToOp(event)
	filename := filepath.Base(event.Name)

	for _, w := range watchers {
		if !hasEvent(w, eventStr) {
			log.Debug().
				Str("event", event.Op.String()).
				Str("file_name", event.Name).
				Msg("no event match found for event")
			continue
		}

		if !matchesPattern(w, filename) {
			log.Debug().
				Str("event", event.Op.String()).
				Str("file_name", event.Name).
				Msg("no file match found for event")
			continue
		}

		cmdstr, err := quicktmpl.Render(w.Exec, quicktmpl.Data{
			"Path": event.Name,
			"Vars": wh.cfg.Vars,
		})
		if err != nil {
			continue
		}

		cmd := exec.Command(wh.cfg.Shell, wh.cfg.ShellCmd, cmdstr)
		cmd.Stdout = log.Logger
		cmd.Stderr = log.Logger

		if err := cmd.Run(); err != nil {
			log.Error().Err(err).Msg("failed to run command")
		}
	}
}

func hasEvent(w config.Watcher, eventStr string) bool {
	for _, e := range w.Events {
		if e == eventStr {
			return true
		}
	}
	return false
}

func matchesPattern(w config.Watcher, filename string) bool {
	for _, d := range w.Match {
		match, err := doublestar.Match(d, filename)
		if err != nil {
			log.Error().Err(err).Msg("failed to match")
			return false
		}

		if match {
			log.Debug().
				Str("file_name", filename).
				Str("pattern", d).
				Msg("match")
			return true
		}

		log.Debug().
			Str("file_name", filename).
			Str("pattern", d).
			Msg("no match")
	}
	return false
}
