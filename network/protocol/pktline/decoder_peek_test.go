package pktline_test

import (
	"strings"
	"testing"

	"lindenii.org/go/furgit/network/protocol/pktline"
)

func TestDecoderPeek(t *testing.T) {
	t.Parallel()

	dec := pktline.NewDecoder(strings.NewReader("0005x0000"), pktline.ReadOptions{})

	f, err := dec.PeekFrame()
	if err != nil {
		t.Fatalf("PeekFrame: %v", err)
	}

	if f.Type != pktline.PacketData || string(f.Payload) != "x" {
		t.Fatalf("peek frame = %#v", f)
	}

	f, err = dec.ReadFrame()
	if err != nil {
		t.Fatalf("ReadFrame: %v", err)
	}

	if f.Type != pktline.PacketData || string(f.Payload) != "x" {
		t.Fatalf("read frame = %#v", f)
	}
}
