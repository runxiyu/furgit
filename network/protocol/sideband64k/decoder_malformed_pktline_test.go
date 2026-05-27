package sideband64k_test

import (
	"errors"
	"strings"
	"testing"

	"lindenii.org/go/furgit/network/protocol/pktline"
	"lindenii.org/go/furgit/network/protocol/sideband64k"
)

func TestDecoderInvalid0003(t *testing.T) {
	t.Parallel()

	dec := sideband64k.NewDecoder(strings.NewReader("0003"), sideband64k.ReadOptions{})
	_, err := dec.ReadFrame()

	if _, ok := errors.AsType[*pktline.ProtocolError](err); !ok {
		t.Fatalf("got err %v, want pktline.ProtocolError", err)
	}
}

func TestDecoderRejectsOverMaximumLength(t *testing.T) {
	t.Parallel()

	dec := sideband64k.NewDecoder(strings.NewReader("fffe"), sideband64k.ReadOptions{})
	_, err := dec.ReadFrame()

	if _, ok := errors.AsType[*pktline.ProtocolError](err); !ok {
		t.Fatalf("got err %v, want pktline.ProtocolError", err)
	}
}
