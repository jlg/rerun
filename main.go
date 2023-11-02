package main

import (
	"flag"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/jlg/rerun/exec"
)

const blEnvName = `RERUN_BLACKLIST`
const blDefault = `
4913
\..*
.*\.sw[px]
.*~
`

type multiFlag []string

func (f *multiFlag) String() string {
	return fmt.Sprint(*f)
}

func (f *multiFlag) Set(value string) error {
	*f = append(*f, value)
	return nil
}

type FilterPaths []*regexp.Regexp

func (f *FilterPaths) mustRegexp(blacklist, sep string) {
	bl := strings.Split(strings.Trim(blacklist, sep), sep)
	for _, p := range bl {
		re := regexp.MustCompile(`^` + p + `$`)
		*f = append(*f, re)
	}
}

func (f FilterPaths) Match(path string) bool {
	if f == nil {
		return false
	}
	path = filepath.Base(path)
	for _, re := range f {
		if re.MatchString(path) {
			return true
		}
	}
	return false
}

func main() {
	var (
		// Flags
		watchDirs multiFlag
		waitFor   time.Duration
		maxFiles  int
		clearTerm bool
	)

	flag.Usage = func() {
		fmt.Fprintf(flag.CommandLine.Output(), "Usage of %s:\n", os.Args[0])
		fmt.Fprintf(flag.CommandLine.Output(), "%s: [-w 100ms] [-s 7] [-c] [-d <dir> -d ...] <command> [commandargs ...]\n", os.Args[0])
		flag.PrintDefaults()
	}
	flag.DurationVar(&waitFor, "w", 100*time.Millisecond, "delay before triggering command")
	flag.IntVar(&maxFiles, "s", 42, "soft max number of files to be passed to command, 0 disables passing")
	flag.BoolVar(&clearTerm, "c", false, "clear terminal before each command")
	flag.Var(&watchDirs, "d", "directories to watch, multiple use of flag allowed (default .)")
	flag.Parse()
	skipArgs := maxFiles == 0

	if len(watchDirs) == 0 {
		watchDirs = append(watchDirs, ".")
	}
	args := flag.Args()
	if len(args) < 1 {
		flag.Usage()
		os.Exit(2)
	}
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, os.Interrupt)

	filterPaths := FilterPaths{}
	filterPaths.mustRegexp(blDefault, "\n")
	filterPaths.mustRegexp(os.Getenv(blEnvName), ":")

	modified, cleanup, err := watcher(watchDirs, filterPaths)
	if err != nil {
		slog.Error("error starting watcher: %w", err)
	}
	defer cleanup()
	go func() {
		<-sigs
		cleanup()
	}()

	files := buffer(modified, maxFiles, waitFor)
	partialCommand := exec.NewPartial(args...)
	// Blocks main goroutine
	executor(files, partialCommand, skipArgs, clearTerm)
	slog.Info("done")
}
