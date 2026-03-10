package sideband64k_test

import (
	"strings"
	"testing"

	"codeberg.org/lindenii/furgit/protocol/sideband64k"
)

func TestDecoderPeek(t *testing.T) {
	t.Parallel()

	dec := sideband64k.NewDecoder(strings.NewReader("0006\x01x0000"), sideband64k.ReadOptions{})

	f, err := dec.PeekFrame()
	if err != nil {
		t.Fatalf("PeekFrame: %v", err)
	}

	if f.Type != sideband64k.FrameData || string(f.Payload) != "x" {
		t.Fatalf("peek frame = %#v", f)
	}

	f.Payload[0] = 'y'

	f, err = dec.ReadFrame()
	if err != nil {
		t.Fatalf("ReadFrame: %v", err)
	}

	if f.Type != sideband64k.FrameData || string(f.Payload) != "x" {
		t.Fatalf("read frame = %#v", f)
	}
}
