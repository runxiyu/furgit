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

	if err := enc.WriteData([]byte("hello")); err != nil {
		t.Fatalf("WriteData: %v", err)
	}
	if err := enc.WriteFlush(); err != nil {
		t.Fatalf("WriteFlush: %v", err)
	}
	if out.Len() != 0 {
		t.Fatalf("WriteFlush should not flush I/O, got %q", out.String())
	}
	if err := enc.FlushIO(); err != nil {
		t.Fatalf("FlushIO: %v", err)
	}
	if got, want := out.String(), "000a\x01hello0000"; got != want {
		t.Fatalf("got %q, want %q", got, want)
	}

	out.Reset()
	bw = bufio.NewWriter(&out)
	enc = sideband64k.NewEncoder(bw)

	if err := enc.WriteData([]byte("yo")); err != nil {
		t.Fatalf("WriteData: %v", err)
	}
	if err := enc.WriteFlushAndFlushIO(); err != nil {
		t.Fatalf("WriteFlushAndFlushIO: %v", err)
	}
	if got, want := out.String(), "0007\x01yo0000"; got != want {
		t.Fatalf("got %q, want %q", got, want)
	}
}
