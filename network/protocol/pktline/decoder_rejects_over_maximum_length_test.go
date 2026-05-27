package pktline_test

import (
	"errors"
	"strings"
	"testing"

	"lindenii.org/go/furgit/network/protocol/pktline"
)

func TestDecoderRejectsOverMaximumLength(t *testing.T) {
	t.Parallel()

	dec := pktline.NewDecoder(strings.NewReader("fffe"), pktline.ReadOptions{})
	dec.SetMaxData(70000)

	_, err := dec.ReadFrame()

	if _, ok := errors.AsType[*pktline.ProtocolError](err); !ok {
		t.Fatalf("got err %v, want ProtocolError", err)
	}
}
