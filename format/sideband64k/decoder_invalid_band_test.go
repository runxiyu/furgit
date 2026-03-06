package sideband64k_test

import (
	"errors"
	"strings"
	"testing"

	"codeberg.org/lindenii/furgit/format/sideband64k"
)

func TestDecoderInvalidBand(t *testing.T) {
	t.Parallel()

	dec := sideband64k.NewDecoder(strings.NewReader("0005\x04"), sideband64k.ReadOptions{})
	_, err := dec.ReadFrame()

	var pe *sideband64k.ProtocolError
	if !errors.As(err, &pe) {
		t.Fatalf("got err %v, want ProtocolError", err)
	}
}
