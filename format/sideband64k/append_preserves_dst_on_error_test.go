package sideband64k_test

import (
	"bytes"
	"errors"
	"testing"

	"codeberg.org/lindenii/furgit/format/sideband64k"
)

func TestAppendBandPreservesDstOnError(t *testing.T) {
	t.Parallel()

	orig := []byte("seed")
	dst := append([]byte(nil), orig...)

	out, err := sideband64k.AppendBand(dst, 4, []byte("x"))
	if !errors.Is(err, sideband64k.ErrInvalidBand) {
		t.Fatalf("got err %v, want ErrInvalidBand", err)
	}

	if !bytes.Equal(out, orig) {
		t.Fatalf("got %q, want %q", string(out), string(orig))
	}

	out, err = sideband64k.AppendData(dst, bytes.Repeat([]byte{'x'}, sideband64k.DataMax+1))
	if !errors.Is(err, sideband64k.ErrTooLarge) {
		t.Fatalf("got err %v, want ErrTooLarge", err)
	}

	if !bytes.Equal(out, orig) {
		t.Fatalf("got %q, want %q", string(out), string(orig))
	}
}
