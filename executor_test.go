package main

import (
	"errors"
	"io"
	"log"
	"reflect"
	"testing"
)

type FakeCmd struct {
	Calls [][]string
	err   error
}

func (c *FakeCmd) Run(args ...string) error {
	c.Calls = append(c.Calls, args)
	return c.err
}

func TestExecutor(t *testing.T) {
	log.SetOutput(io.Discard) // Disable logging for tests

	data := []struct {
		name   string
		args   []string
		skip   bool
		repeat int
		called []string
		cmdErr error
	}{
		{
			name:   "executor calls command with two args",
			args:   []string{"file1", "file2"},
			skip:   false,
			repeat: 1,
			called: []string{"file1", "file2"},
		},
		{
			name:   "executor calls command twice with args",
			args:   []string{"file1", "file2"},
			skip:   false,
			repeat: 2,
			called: []string{"file1", "file2"},
		},
		{
			name:   "executor skips args to command when skipArgs==ture",
			args:   []string{"file1", "file2"},
			skip:   true,
			repeat: 2,
			called: []string{},
		},
		{
			name:   "executor handles command error",
			args:   []string{"file1", "file2"},
			skip:   true,
			repeat: 2,
			called: []string{},
			cmdErr: errors.New("command error"),
		},
	}
	for _, d := range data {
		t.Run(d.name, func(t *testing.T) {
			fc := FakeCmd{err: d.cmdErr}
			ch := make(chan []string, d.repeat)

			expectedCalls := [][]string{}
			for i := 0; i < d.repeat; i++ {
				ch <- d.args
				expectedCalls = append(expectedCalls, d.called)
			}
			close(ch)

			executor(ch, &fc, d.skip, false)

			if len(fc.Calls) != d.repeat {
				t.Fatalf("expected %d calls to FakeCmd.Run, got %d", d.repeat, len(fc.Calls))
			}
			if !reflect.DeepEqual(expectedCalls, fc.Calls) {
				t.Fatalf("expected FakeCmd.Run to be called with %v, got %v", expectedCalls, fc.Calls)
			}
		})
	}
}
