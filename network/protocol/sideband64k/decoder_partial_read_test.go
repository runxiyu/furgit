package sideband64k_test

import (
	"testing"

	"lindenii.org/go/furgit/network/protocol/sideband64k"
)

func TestDecoderHandlesPartialReads(t *testing.T) {
	t.Parallel()

	r := &byteReader{data: []byte("0007\x02ok0000")}
	dec := sideband64k.NewDecoder(r, sideband64k.ReadOptions{})

	f, err := dec.ReadFrame()
	if err != nil {
		t.Fatalf("ReadFrame #1: %v", err)
	}

	if f.Type != sideband64k.FrameProgress || string(f.Payload) != "ok" {
		t.Fatalf("frame #1 = %#v", f)
	}

	f, err = dec.ReadFrame()
	if err != nil {
		t.Fatalf("ReadFrame #2: %v", err)
	}

	if f.Type != sideband64k.FrameFlush {
		t.Fatalf("frame #2 = %#v", f)
	}
}
