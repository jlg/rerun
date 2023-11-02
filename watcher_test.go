package main

import (
	"os"
	"strings"
	"testing"

	"github.com/jlg/rerun/internal"
)

type SuffixFilter string

func (f SuffixFilter) Match(s string) bool {
	return strings.HasSuffix(s, string(f))
}

func createWatchedDirs(t testing.TB, n int, f Filter) ([]string, <-chan string) {
	var dirs []string
	t.Helper()
	for i := 0; i < n; i++ {
		dirs = append(dirs, t.TempDir())
	}

	ch, cleanup, err := watcher(dirs, f)
	internal.NoErr(t, err)
	t.Cleanup(cleanup)
	return dirs, ch
}

func TestWatcher(t *testing.T) {
	t.Run("watcher with empty dirs slice", func(t *testing.T) {
		dirs := []string{}

		_, _, err := watcher(dirs, nil)
		if err == nil {
			t.Fatal("expected error, got nil")
		}
	})
	t.Run("watcher return channel, chanel closes after cleanup", func(t *testing.T) {
		dir := t.TempDir()
		dirs := []string{dir}

		ch, cleanup, err := watcher(dirs, nil)
		internal.NoErr(t, err)

		internal.ExpectEmpty(t, ch)

		cleanup()
		internal.ExpectClosed(t, ch)
	})

	t.Run("create and write to file triggers twice", func(t *testing.T) {
		dirs, ch := createWatchedDirs(t, 1, nil)

		f, err := os.CreateTemp(dirs[0], "testfile")
		internal.NoErr(t, err)

		mod := internal.ExpectRead(t, ch)
		if mod != f.Name() {
			t.Fatalf("expected: %s got: %s", f.Name(), mod)
		}

		f.WriteString("testdata")
		f.Close()

		mod = internal.ExpectRead(t, ch)
		if mod != f.Name() {
			t.Fatalf("expected: %s got: %s", f.Name(), mod)
		}

		os.Remove(f.Name())
		internal.ExpectEmpty(t, ch)
	})

	t.Run("create files triggers in two watched dirs", func(t *testing.T) {
		dirs, ch := createWatchedDirs(t, 2, nil)

		for _, dir := range dirs {
			f, err := os.CreateTemp(dir, "testfile")
			internal.NoErr(t, err)

			mod := internal.ExpectRead(t, ch)
			if mod != f.Name() {
				t.Fatalf("expected: %s got: %s", f.Name(), mod)
			}
			f.Close()
		}
	})

	t.Run("create two files triggers with both", func(t *testing.T) {
		dir, ch := createWatchedDirs(t, 1, nil)

		for _, fname := range []string{"file1", "file2"} {
			f, err := os.CreateTemp(dir[0], fname)
			internal.NoErr(t, err)

			mod := internal.ExpectRead(t, ch)
			if mod != f.Name() {
				t.Fatalf("expected: %s got: %s", f.Name(), mod)
			}
		}
	})
	t.Run("create two files skips matching path", func(t *testing.T) {
		dirs, ch := createWatchedDirs(t, 1, SuffixFilter(".no"))

		_, err := os.CreateTemp(dirs[0], "file*.no")
		internal.NoErr(t, err)
		internal.ExpectEmpty(t, ch)

		f, err := os.CreateTemp(dirs[0], "file*.yes")
		internal.NoErr(t, err)

		mod := internal.ExpectRead(t, ch)
		if mod != f.Name() {
			t.Fatalf("expected: %s got: %s", f.Name(), mod)
		}
		internal.ExpectEmpty(t, ch)
	})
}
