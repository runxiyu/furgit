package pktline_test

import (
	"errors"
	"testing"

	"codeberg.org/lindenii/furgit/network/protocol/pktline"
)

func TestEncodeLengthHeader(t *testing.T) {
	t.Parallel()

	var hdr [4]byte

	err := pktline.EncodeLengthHeader(&hdr, 4)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if got := string(hdr[:]); got != "0004" {
		t.Fatalf("got %q, want %q", got, "0004")
	}

	err = pktline.EncodeLengthHeader(&hdr, pktline.LargePacketMax+1)
	if !errors.Is(err, pktline.ErrInvalidLength) {
		t.Fatalf("got err %v, want ErrInvalidLength", err)
	}
}
