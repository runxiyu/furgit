package pktline_test

import (
	"errors"
	"testing"
	"codeberg.org/lindenii/furgit/format/pktline"
)

func TestParseLengthHeader(t *testing.T) {
	t.Parallel()

	n, err := pktline.ParseLengthHeader([4]byte{'0', '0', '0', '4'})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if n != 4 {
		t.Fatalf("got %d, want 4", n)
	}

	_, err = pktline.ParseLengthHeader([4]byte{'0', '0', '0', 'x'})
	if !errors.Is(err, pktline.ErrInvalidLength) {
		t.Fatalf("got err %v, want ErrInvalidLength", err)
	}
}

