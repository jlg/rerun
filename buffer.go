package main

import (
	"log/slog"
	"time"
)

// Buffer multiple file modifications into buffSize, allow waitFor time for buffering.
// Forward deduplicated files after buffSize unique modifications reached or waitFor elapsed.
func buffer(files <-chan string, buffSize int, waitFor time.Duration) <-chan []string {
	args := make(chan []string)
	modified := make(map[string]bool)
	// Minimum size of 2 to allow for de-duplicaiton (repeated mod to same file).
	// Change to only one file has falls back on timer.
	buffSize = max(buffSize, 2)

	go func() {
		timer := time.NewTimer(waitFor)
		timer.Stop()
		defer timer.Stop()
		defer close(args)
	loop:
		for {
			select {
			case path, ok := <-files:
				if !ok { // files closed
					break loop
				}
				modified[path] = true
				if len(modified) < buffSize {
					// Allow waitFor to get more files
					timer.Reset(waitFor)
					continue
				}
				slog.Debug("runner trigerred by full buffer")

			case <-timer.C:
				if len(modified) == 0 {
					// This should not happen.
					timer.Stop()
					continue
				}
				slog.Debug("runner trigerred by delay")
			}
			timer.Stop()
			// Send files because of full buffer or timer firing with non empty buffer
			args <- drainModified(modified)
		}
	}()

	return args
}

// Drain modified set into output slice. Empties modified map.
func drainModified(modified map[string]bool) []string {
	files := make([]string, 0, len(modified))
	for k := range modified {
		files = append(files, k)
		delete(modified, k)
	}
	return files
}
