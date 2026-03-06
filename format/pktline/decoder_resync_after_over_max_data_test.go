package pktline_test

import (
	"bufio"
	"bytes"
	"errors"
	"testing"
	"codeberg.org/lindenii/furgit/format/pktline"
)

func TestDecoderResyncAfterOverMaxData(t *testing.T) {
	t.Parallel()

	var b bytes.Buffer

	bw := bufio.NewWriter(&b)
	enc := pktline.NewEncoder(bw)

	err := enc.WriteData([]byte("abcd"))
	if err != nil {
		t.Fatalf("WriteData #1: %v", err)
	}

	err = enc.WriteData([]byte("z"))
	if err != nil {
		t.Fatalf("WriteData #2: %v", err)
	}

	err = enc.FlushIO()
	if err != nil {
		t.Fatalf("FlushIO: %v", err)
	}

	dec := pktline.NewDecoder(bytes.NewReader(b.Bytes()), pktline.ReadOptions{})
	dec.SetMaxData(1)

	_, err = dec.ReadFrame()
	if !errors.Is(err, pktline.ErrTooLarge) {
		t.Fatalf("got err %v, want ErrTooLarge", err)
	}

	f, err := dec.ReadFrame()
	if err != nil {
		t.Fatalf("ReadFrame #2: %v", err)
	}

	if f.Type != pktline.PacketData || string(f.Payload) != "z" {
		t.Fatalf("got frame %#v, want data z", f)
	}
}

