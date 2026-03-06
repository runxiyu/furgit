package pktline_test

import (
	"bufio"
	"bytes"
	"errors"
	"testing"

	"codeberg.org/lindenii/furgit/format/pktline"
)

func TestEncoderSetMaxDataCannotExceedWireLimit(t *testing.T) {
	t.Parallel()

	var out bytes.Buffer

	bw := bufio.NewWriter(&out)

	enc := pktline.NewEncoder(bw)
	enc.SetMaxData(pktline.LargePacketDataMax + 100)

	err := enc.WriteData(bytes.Repeat([]byte{'x'}, pktline.LargePacketDataMax+1))
	if !errors.Is(err, pktline.ErrTooLarge) {
		t.Fatalf("got err %v, want ErrTooLarge", err)
	}
}
