package internal

import (
	"testing"
	"time"
)

func ExpectClosed[T any](t testing.TB, c <-chan T) bool {
	t.Helper()
	select {
	case v, ok := <-c:
		if !ok {
			return true
		}
		t.Errorf("expected closed channel: read %v", v)
	case <-time.After(10 * time.Millisecond):
	}
	t.Error("expected closed channel: empty")
	return false
}

func ExpectRead[T any](t testing.TB, c <-chan T) T {
	t.Helper()
	select {
	case v, ok := <-c:
		if !ok {
			t.Error("expected to read form channel: closed")
		}
		return v
	case <-time.After(10 * time.Millisecond):
	}
	t.Error("expected to read form channel: empty")
	var v T
	return v
}

func ExpectEmpty[T any](t testing.TB, c <-chan T) bool {
	t.Helper()
	select {
	case v, ok := <-c:
		if !ok {
			t.Error("expected empty channel: closed")
		}
		t.Errorf("expected closed channel: read %v", v)
		return false
	case <-time.After(10 * time.Millisecond):
	}
	return true
}

func NoErr(t testing.TB, err error) {
	t.Helper()
	if err != nil {
		t.Fatalf("got unexpected error: %v", err)
	}
}
