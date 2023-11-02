package main

import (
	"io"
	"log"
	"reflect"
	"slices"
	"testing"
	"time"

	"github.com/jlg/rerun/internal"
)

func TestBuffer(t *testing.T) {
	log.SetOutput(io.Discard) // Disable logging for tests

	t.Run("buffer passes file after wait", func(t *testing.T) {
		in := make(chan string, 5)
		buffSize := 5
		waitFor := 2 * time.Millisecond

		out := buffer(in, buffSize, waitFor)
		in <- "file1"
		expected := []string{"file1"}

		args := internal.ExpectRead(t, out)
		if !reflect.DeepEqual(args, expected) {
			t.Fatalf("expected %v, got: %v", expected, args)
		}
		internal.ExpectEmpty(t, out)
		close(in)
		internal.ExpectClosed(t, out)
	})

	t.Run("buffer passes two files after wait", func(t *testing.T) {
		in := make(chan string, 5)
		buffSize := 5
		waitFor := 2 * time.Millisecond

		out := buffer(in, buffSize, waitFor)
		in <- "file1"
		in <- "file2"
		expected := []string{"file1", "file2"}

		args := internal.ExpectRead(t, out)
		slices.Sort(args)
		if !reflect.DeepEqual(args, expected) {
			t.Fatalf("expected %v, got: %v", expected, args)
		}
		internal.ExpectEmpty(t, out)
		close(in)
		internal.ExpectClosed(t, out)
	})

	t.Run("buffer de-duplicates files", func(t *testing.T) {
		in := make(chan string, 5)
		buffSize := 5
		waitFor := 2 * time.Millisecond

		out := buffer(in, buffSize, waitFor)
		in <- "file1"
		in <- "file2"
		in <- "file2"
		in <- "file1"
		in <- "file1"
		in <- "file2"
		expected := []string{"file1", "file2"}

		args := internal.ExpectRead(t, out)
		slices.Sort(args)
		if !reflect.DeepEqual(args, expected) {
			t.Fatalf("expected %v, got: %v", expected, args)
		}
		internal.ExpectEmpty(t, out)
		close(in)
		internal.ExpectClosed(t, out)
	})

	t.Run("full buffer passes files before wait", func(t *testing.T) {
		in := make(chan string, 2)
		buffSize := 2
		waitFor := 20 * time.Millisecond

		out := buffer(in, buffSize, waitFor)
		in <- "file1"
		in <- "file2"
		expected := []string{"file1", "file2"}

		args := internal.ExpectRead(t, out)
		slices.Sort(args)
		if !reflect.DeepEqual(args, expected) {
			t.Fatalf("expected %v, got: %v", expected, args)
		}
		internal.ExpectEmpty(t, out)
		close(in)
		internal.ExpectClosed(t, out)
	})

	t.Run("buffer overflow passes buffSize files", func(t *testing.T) {
		in := make(chan string, 2)
		buffSize := 2
		waitFor := 5 * time.Millisecond

		out := buffer(in, buffSize, waitFor)
		internal.ExpectEmpty(t, out)
		in <- "file1"
		in <- "file2"
		// Buffer overflow
		in <- "file3"
		expected := []string{"file1", "file2", "file3"}

		args := internal.ExpectRead(t, out)
		slices.Sort(args)
		if !reflect.DeepEqual(args, expected[:2]) {
			t.Fatalf("expected %v, got: %v", expected[:2], args)
		}

		args = internal.ExpectRead(t, out)
		if !reflect.DeepEqual(args, expected[2:]) {
			t.Fatalf("expected %v, got: %v", expected[2:], args)
		}

		internal.ExpectEmpty(t, out)
		close(in)
		internal.ExpectClosed(t, out)
	})
}
