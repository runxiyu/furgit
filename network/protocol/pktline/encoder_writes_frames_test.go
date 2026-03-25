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

	err = enc.WriteFlush()
	if err != nil {
		t.Fatalf("WriteFlush: %v", err)
	}

	err = enc.WriteDelim()
	if err != nil {
		t.Fatalf("WriteDelim: %v", err)
	}

	err = enc.WriteResponseEnd()
	if err != nil {
		t.Fatalf("WriteResponseEnd: %v", err)
	}

	err = enc.FlushIO()
	if err != nil {
		t.Fatalf("FlushIO: %v", err)
	}

	got := b.String()

	want := "0006hi000000010002"
	if got != want {
		t.Fatalf("got %q, want %q", got, want)
	}
}
