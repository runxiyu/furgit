package sideband64k_test

import (
	"bufio"
	"bytes"
	"errors"
	"testing"

	"codeberg.org/lindenii/furgit/network/protocol/sideband64k"
)

func TestDecoderResyncAfterOverMaxData(t *testing.T) {
	t.Parallel()

	var b bytes.Buffer

	bw := bufio.NewWriter(&b)
	enc := sideband64k.NewEncoder(bw)

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

	dec := sideband64k.NewDecoder(bytes.NewReader(b.Bytes()), sideband64k.ReadOptions{})
	dec.SetMaxData(1)

	_, err = dec.ReadFrame()
	if !errors.Is(err, sideband64k.ErrTooLarge) {
		t.Fatalf("got err %v, want ErrTooLarge", err)
	}

	f, err := dec.ReadFrame()
	if err != nil {
		t.Fatalf("ReadFrame #2: %v", err)
	}

	if f.Type != sideband64k.FrameData || string(f.Payload) != "z" {
		t.Fatalf("got frame %#v, want data z", f)
	}
}
