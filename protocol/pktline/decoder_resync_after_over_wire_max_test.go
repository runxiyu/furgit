package pktline_test

import (
	"bytes"
	"errors"
	"testing"

	"codeberg.org/lindenii/furgit/protocol/pktline"
)

func TestDecoderResyncAfterOverWireMax(t *testing.T) {
	t.Parallel()

	var b bytes.Buffer

	_, _ = b.WriteString("ffff")
	_, _ = b.Write(bytes.Repeat([]byte{'a'}, 65531))
	_, _ = b.WriteString("0005z")

	dec := pktline.NewDecoder(bytes.NewReader(b.Bytes()), pktline.ReadOptions{})
	dec.SetMaxData(70000)

	_, err := dec.ReadFrame()

	if _, ok := errors.AsType[*pktline.ProtocolError](err); !ok {
		t.Fatalf("got err %v, want ProtocolError", err)
	}

	f, err := dec.ReadFrame()
	if err != nil {
		t.Fatalf("ReadFrame #2: %v", err)
	}

	if f.Type != pktline.PacketData || string(f.Payload) != "z" {
		t.Fatalf("got frame %#v, want data z", f)
	}
}
