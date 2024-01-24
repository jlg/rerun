package main

import (
	"bufio"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/jlg/rerun/internal"
)

func writeTemp(tempDir, pattern, data string) string {
	f, _ := os.CreateTemp(tempDir, pattern)
	defer f.Close()
	f.Write([]byte(data))
	return f.Name()
}

func expectTokens(t testing.TB, s *bufio.Scanner, expected ...string) {
	t.Helper()
	s.Scan()
	line := strings.Split(s.Text(), " ")
	slices.Sort(line)
	slices.Sort(expected)
	if !reflect.DeepEqual(line, expected) {
		t.Errorf("got: '%s', expected '%s'", line, expected)
	}
}

func TestCommandLine(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integraton test in -short mode")
	}
	tempDir := t.TempDir()
	binPath := filepath.Join(tempDir, "rerun.bin")
	softTimeout := 300 * time.Millisecond
	hardTimeout := 400 * time.Millisecond

	build := exec.Command("go", "build", "-o", binPath)
	err := build.Run()
	internal.NoErr(t, err)

	r, w := io.Pipe()
	rerun := exec.Command(binPath, "-w=20ms", "-s=2", "-d", tempDir, "echo")
	rerun.Stdout = w
	err = rerun.Start()

	time.AfterFunc(softTimeout, func() {
		w.Close()
	})
	time.AfterFunc(hardTimeout, func() {
		rerun.Process.Signal(os.Kill)
	})

	internal.NoErr(t, err)

	time.Sleep(20 * time.Millisecond) // FIXME make sure it is up and running

	scanner := bufio.NewScanner(r)

	f1 := writeTemp(tempDir, "file1-*.no", "1234567890")

	expectTokens(t, scanner, f1)

	f2 := writeTemp(tempDir, "file2-*.yes", "12345")
	f3 := writeTemp(tempDir, "file3-*.maybe", "")

	expectTokens(t, scanner, f2, f3)

	rerun.Process.Signal(os.Interrupt) // handles ^C with no err
	err = rerun.Wait()
	if err != nil {
		t.Fatalf("%s terminated by: signal %v", rerun.Path, err)
	}
}

func TestRerun(t *testing.T) {
	stdout := os.Stdout
	defer func() {
		os.Stdout = stdout
	}()

	tempDir := t.TempDir()
	rerun := Rerun{
		watchDirs:   []string{tempDir},
		maxFiles:    2,
		waitFor:     20 * time.Millisecond,
		commandArgs: []string{"echo"},
	}
	rerun.filterPaths.MustRegexp(`.*\.skip`, "\n")

	r, w, err := os.Pipe()
	internal.NoErr(t, err)
	os.Stdout = w

	cancel, err := rerun.Start()
	internal.NoErr(t, err)

	f1 := writeTemp(tempDir, "file1-*.no", "1234567890")

	scanner := bufio.NewScanner(r)
	expectTokens(t, scanner, f1)

	f2 := writeTemp(tempDir, "file2-*.yes", "12345")
	writeTemp(tempDir, "file3-*.skip", "123")
	f3 := writeTemp(tempDir, "file4-*.maybe", "")

	expectTokens(t, scanner, f2, f3)

	cancel()
	err = rerun.Wait()
	internal.NoErr(t, err)
}
