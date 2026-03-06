package sideband64k_test

import (
	"bufio"
	"bytes"
	"testing"

	"codeberg.org/lindenii/furgit/format/sideband64k"
)

func TestEncoderWritesFrames(t *testing.T) {
	t.Parallel()

	var b bytes.Buffer

	bw := bufio.NewWriter(&b)
	enc := sideband64k.NewEncoder(bw)

	if err := enc.WriteData([]byte("hi")); err != nil {
		t.Fatalf("WriteData: %v", err)
	}
	if err := enc.WriteProgress([]byte("ok")); err != nil {
		t.Fatalf("WriteProgress: %v", err)
	}
	if err := enc.WriteError([]byte("no")); err != nil {
		t.Fatalf("WriteError: %v", err)
	}
	if err := enc.WriteFlush(); err != nil {
		t.Fatalf("WriteFlush: %v", err)
	}
	if err := enc.WriteDelim(); err != nil {
		t.Fatalf("WriteDelim: %v", err)
	}
	if err := enc.WriteResponseEnd(); err != nil {
		t.Fatalf("WriteResponseEnd: %v", err)
	}
	if err := enc.FlushIO(); err != nil {
		t.Fatalf("FlushIO: %v", err)
	}

	want := "0007\x01hi0007\x02ok0007\x03no000000010002"
	if got := b.String(); got != want {
		t.Fatalf("got %q, want %q", got, want)
	}
}
