package pktline_test

import (
	"bufio"
	"bytes"
	"testing"

	"codeberg.org/lindenii/furgit/format/pktline"
)

func TestEncoderBufferedFlushAndFFlush(t *testing.T) {
	t.Parallel()

	var out bytes.Buffer

	bw := bufio.NewWriter(&out)
	enc := pktline.NewEncoder(bw)

	err := enc.WriteData([]byte("x"))
	if err != nil {
		t.Fatalf("WriteData: %v", err)
	}

	if out.Len() != 0 {
		t.Fatalf("unexpected immediate output: %q", out.String())
	}

	err = enc.FlushIO()
	if err != nil {
		t.Fatalf("FlushIO: %v", err)
	}

	if out.String() != "0005x" {
		t.Fatalf("got %q, want %q", out.String(), "0005x")
	}

	out.Reset()
	bw = bufio.NewWriter(&out)

	enc = pktline.NewEncoder(bw)

	err = enc.WriteFlushAndFlushIO()
	if err != nil {
		t.Fatalf("WriteFlushAndFlushIO: %v", err)
	}

	if out.String() != "0000" {
		t.Fatalf("got %q, want %q", out.String(), "0000")
	}
}
