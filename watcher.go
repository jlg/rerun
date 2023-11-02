package main

import (
	"errors"
	"fmt"
	"log/slog"

	"github.com/fsnotify/fsnotify"
)

type Filter interface {
	Match(s string) bool
}

// Watches for file Create and Write events in paths and filters out blacklisted filenames.
// Pass modified filename.
func watcher(paths []string, filter Filter) (<-chan string, func(), error) {
	if len(paths) == 0 {
		return nil, nil, errors.New("no paths to watch")
	}
	fswatcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, nil, fmt.Errorf("creating watcher: %w", err)
	}
	cleanup := func() {
		fswatcher.Close()
	}

	// Start listening for events.
	files := watchLoop(fswatcher, filter)

	// Add a path.
	for _, path := range paths {
		err = fswatcher.Add(path)
		if err != nil {
			cleanup()
			return nil, nil, fmt.Errorf("adding path '%v': %w", path, err)
		}
	}
	return files, cleanup, nil
}

func watchLoop(fswatcher *fsnotify.Watcher, filter Filter) <-chan string {
	files := make(chan string)

	go func() {
		defer close(files)
	loop:
		for {
			select {
			case e, ok := <-fswatcher.Events:
				if !ok {
					break loop
				}
				if !e.Has(fsnotify.Write) && !e.Has(fsnotify.Create) {
					continue
				}
				if filter != nil && filter.Match(e.Name) {
					slog.Debug("skipped", "filename", e.Name)
					continue
				}
				slog.Debug("modified", "filename", e.Name)
				files <- e.Name
			case err, ok := <-fswatcher.Errors:
				if !ok {
					break loop
				}
				slog.Error("watcher error", "error", err)
			}
		}
	}()

	return files
}
