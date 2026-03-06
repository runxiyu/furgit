package sideband64k_test

import (
	"bytes"
	"errors"
	"testing"

	"codeberg.org/lindenii/furgit/format/pktline"
	"codeberg.org/lindenii/furgit/format/sideband64k"
)

func TestDecoderResyncAfterOverWireMax(t *testing.T) {
	t.Parallel()

	var b bytes.Buffer

	_, _ = b.WriteString("ffff")
	_, _ = b.Write(bytes.Repeat([]byte{'a'}, 65531))
	_, _ = b.WriteString("0006\x01z")

	dec := sideband64k.NewDecoder(bytes.NewReader(b.Bytes()), sideband64k.ReadOptions{})

	_, err := dec.ReadFrame()

	if _, ok := errors.AsType[*pktline.ProtocolError](err); !ok {
		t.Fatalf("got err %v, want pktline.ProtocolError", err)
	}

	f, err := dec.ReadFrame()
	if err != nil {
		t.Fatalf("ReadFrame #2: %v", err)
	}

	if f.Type != sideband64k.FrameData || string(f.Payload) != "z" {
		t.Fatalf("got frame %#v, want data z", f)
	}
}
