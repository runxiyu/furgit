package sideband64k_test

import (
	"errors"
	"io"
	"strings"
	"testing"

	"lindenii.org/go/furgit/network/protocol/sideband64k"
)

func TestDecoderUnexpectedEOF(t *testing.T) {
	t.Parallel()

	dec := sideband64k.NewDecoder(strings.NewReader("0006\x01"), sideband64k.ReadOptions{})

	_, err := dec.ReadFrame()
	if !errors.Is(err, io.ErrUnexpectedEOF) {
		t.Fatalf("got err %v, want io.ErrUnexpectedEOF", err)
	}
}
