package sideband64k_test

import (
	"testing"

	"lindenii.org/go/furgit/network/protocol/sideband64k"
)

func TestAppendHelpers(t *testing.T) {
	t.Parallel()

	out, err := sideband64k.AppendData(nil, []byte("a"))
	if err != nil {
		t.Fatalf("AppendData: %v", err)
	}

	out, err = sideband64k.AppendProgress(out, []byte("b"))
	if err != nil {
		t.Fatalf("AppendProgress: %v", err)
	}

	out, err = sideband64k.AppendError(out, []byte("c"))
	if err != nil {
		t.Fatalf("AppendError: %v", err)
	}

	if got, want := string(out), "0006\x01a0006\x02b0006\x03c"; got != want {
		t.Fatalf("got %q, want %q", got, want)
	}
}
