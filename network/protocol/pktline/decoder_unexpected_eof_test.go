package pktline_test

import (
	"errors"
	"io"
	"strings"
	"testing"

	"codeberg.org/lindenii/furgit/network/protocol/pktline"
)

func TestDecoderUnexpectedEOF(t *testing.T) {
	t.Parallel()

	dec := pktline.NewDecoder(strings.NewReader("0006a"), pktline.ReadOptions{})

	_, err := dec.ReadFrame()
	if !errors.Is(err, io.ErrUnexpectedEOF) {
		t.Fatalf("got err %v, want io.ErrUnexpectedEOF", err)
	}
}
