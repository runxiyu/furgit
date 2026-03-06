package pktline_test

import (
	"errors"
	"strings"
	"testing"
	"codeberg.org/lindenii/furgit/format/pktline"
)

func TestDecoderRejectsOverMaximumLength(t *testing.T) {
	t.Parallel()

	dec := pktline.NewDecoder(strings.NewReader("fffe"), pktline.ReadOptions{})
	dec.SetMaxData(70000)

	_, err := dec.ReadFrame()

	var pe *pktline.ProtocolError
	if !errors.As(err, &pe) {
		t.Fatalf("got err %v, want ProtocolError", err)
	}
}

