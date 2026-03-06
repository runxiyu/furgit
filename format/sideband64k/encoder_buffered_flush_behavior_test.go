package sideband64k_test

import (
	"bufio"
	"bytes"
	"testing"

	"codeberg.org/lindenii/furgit/format/sideband64k"
)

func TestEncoderBufferedFlushBehavior(t *testing.T) {
	t.Parallel()

	var out bytes.Buffer

	bw := bufio.NewWriter(&out)
	enc := sideband64k.NewEncoder(bw)

	err := enc.WriteData([]byte("hello"))
	if err != nil {
		t.Fatalf("WriteData: %v", err)
	}

	err = enc.WriteFlush()
	if err != nil {
		t.Fatalf("WriteFlush: %v", err)
	}

	if out.Len() != 0 {
		t.Fatalf("WriteFlush should not flush I/O, got %q", out.String())
	}

	err = enc.FlushIO()
	if err != nil {
		t.Fatalf("FlushIO: %v", err)
	}

	if got, want := out.String(), "000a\x01hello0000"; got != want {
		t.Fatalf("got %q, want %q", got, want)
	}

	out.Reset()
	bw = bufio.NewWriter(&out)
	enc = sideband64k.NewEncoder(bw)

	err = enc.WriteData([]byte("yo"))
	if err != nil {
		t.Fatalf("WriteData: %v", err)
	}

	err = enc.WriteFlushAndFlushIO()
	if err != nil {
		t.Fatalf("WriteFlushAndFlushIO: %v", err)
	}

	if got, want := out.String(), "0007\x01yo0000"; got != want {
		t.Fatalf("got %q, want %q", got, want)
	}
}
