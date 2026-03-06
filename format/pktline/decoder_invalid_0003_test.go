package pktline_test

import (
	"errors"
	"strings"
	"testing"
	"codeberg.org/lindenii/furgit/format/pktline"
)

func TestDecoderInvalid0003(t *testing.T) {
	t.Parallel()

	dec := pktline.NewDecoder(strings.NewReader("0003"), pktline.ReadOptions{})
	_, err := dec.ReadFrame()

	var pe *pktline.ProtocolError
	if !errors.As(err, &pe) {
		t.Fatalf("got err %v, want ProtocolError", err)
	}
}

