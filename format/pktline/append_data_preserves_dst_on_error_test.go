package pktline_test

import (
	"bytes"
	"errors"
	"testing"

	"codeberg.org/lindenii/furgit/format/pktline"
)

func TestAppendDataPreservesDstOnError(t *testing.T) {
	t.Parallel()

	orig := []byte("seed")
	dst := append([]byte(nil), orig...)

	out, err := pktline.AppendData(dst, bytes.Repeat([]byte{'x'}, pktline.LargePacketDataMax+1))
	if !errors.Is(err, pktline.ErrTooLarge) {
		t.Fatalf("got err %v, want ErrTooLarge", err)
	}

	if !bytes.Equal(out, orig) {
		t.Fatalf("got %q, want %q", string(out), string(orig))
	}
}
