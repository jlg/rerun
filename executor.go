package main

import (
	"fmt"
	"log/slog"
)

type Command interface {
	Run(args ...string) error
}

// Command executor
//
// Will block until files chan is closed
func executor(files <-chan []string, c Command, skipArgs, clearTerm bool) {
	for args := range files {
		if clearTerm {
			fmt.Print("\033[H\033[2J") // TODO other terminals support
		}
		if skipArgs {
			args = args[:0]
		}
		slog.Info("exec", "command", c, "files", args)
		err := c.Run(args...)
		if err != nil {
			slog.Error("exec failed", "error", err, "command", c, "files", args)
		}
	}
}
