package pktline_test

import (
	"bufio"
	"bytes"
	"testing"

	"codeberg.org/lindenii/furgit/network/protocol/pktline"
)

func TestEncoderWritesFrames(t *testing.T) {
	t.Parallel()

	var b bytes.Buffer

	bw := bufio.NewWriter(&b)

	enc := pktline.NewEncoder(bw)

	err := enc.WriteData([]byte("hi"))
	if err != nil {
		t.Fatalf("WriteData: %v", err)
	}

	err = enc.WriteFlushPacket()
	if err != nil {
		t.Fatalf("WriteFlushPacket: %v", err)
	}

	err = enc.WriteDelimPacket()
	if err != nil {
		t.Fatalf("WriteDelimPacket: %v", err)
	}

	err = enc.WriteResponseEndPacket()
	if err != nil {
		t.Fatalf("WriteResponseEndPacket: %v", err)
	}

	err = enc.Flush()
	if err != nil {
		t.Fatalf("Flush: %v", err)
	}

	got := b.String()

	want := "0006hi000000010002"
	if got != want {
		t.Fatalf("got %q, want %q", got, want)
	}
}
