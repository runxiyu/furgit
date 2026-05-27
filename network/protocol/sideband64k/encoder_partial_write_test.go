package sideband64k_test

import (
	"errors"
	"io"
	"testing"

	"lindenii.org/go/furgit/network/protocol/sideband64k"
)

func TestEncoderHandlesPartialWrites(t *testing.T) {
	t.Parallel()

	dst := &limitWriter{maxPerWrite: 2}
	enc := sideband64k.NewEncoder(dst)

	err := enc.WriteProgress([]byte("abc"))
	if err != nil {
		t.Fatalf("WriteProgress: %v", err)
	}

	err = enc.WriteFlushPacketAndFlush()
	if err != nil {
		t.Fatalf("WriteFlushPacketAndFlush: %v", err)
	}

	if got, want := dst.buf.String(), "0008\x02abc0000"; got != want {
		t.Fatalf("got %q, want %q", got, want)
	}

	if dst.flushes != 1 {
		t.Fatalf("flushes=%d, want 1", dst.flushes)
	}
}

func TestEncoderReturnsShortWrite(t *testing.T) {
	t.Parallel()

	dst := &limitWriter{shortWrite: true}
	enc := sideband64k.NewEncoder(dst)

	err := enc.WriteData([]byte("x"))
	if !errors.Is(err, io.ErrShortWrite) {
		t.Fatalf("got err %v, want io.ErrShortWrite", err)
	}
}
