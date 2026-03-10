package sideband64k_test

import (
	"bufio"
	"bytes"
	"testing"

	"codeberg.org/lindenii/furgit/protocol/sideband64k"
)

func TestEncoderWritesFrames(t *testing.T) {
	t.Parallel()

	var b bytes.Buffer

	bw := bufio.NewWriter(&b)
	enc := sideband64k.NewEncoder(bw)

	err := enc.WriteData([]byte("hi"))
	if err != nil {
		t.Fatalf("WriteData: %v", err)
	}

	err = enc.WriteProgress([]byte("ok"))
	if err != nil {
		t.Fatalf("WriteProgress: %v", err)
	}

	err = enc.WriteError([]byte("no"))
	if err != nil {
		t.Fatalf("WriteError: %v", err)
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

	want := "0007\x01hi0007\x02ok0007\x03no000000010002"
	if got := b.String(); got != want {
		t.Fatalf("got %q, want %q", got, want)
	}
}
