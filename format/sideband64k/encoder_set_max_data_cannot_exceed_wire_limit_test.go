package sideband64k_test

import (
	"bytes"
	"errors"
	"testing"

	"codeberg.org/lindenii/furgit/format/sideband64k"
)

func TestEncoderSetMaxDataCannotExceedWireLimit(t *testing.T) {
	t.Parallel()

	var dst limitWriter

	enc := sideband64k.NewEncoder(&dst)
	enc.SetMaxData(sideband64k.DataMax + 100)

	err := enc.WriteData(bytes.Repeat([]byte{'x'}, sideband64k.DataMax+1))
	if !errors.Is(err, sideband64k.ErrTooLarge) {
		t.Fatalf("got err %v, want ErrTooLarge", err)
	}
}
