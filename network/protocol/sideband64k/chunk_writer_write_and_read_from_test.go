package sideband64k_test

import (
	"bufio"
	"bytes"
	"strings"
	"testing"

	"codeberg.org/lindenii/furgit/network/protocol/sideband64k"
)

func TestChunkWriterWriteAndReadFrom(t *testing.T) {
	t.Parallel()

	var out bytes.Buffer

	bw := bufio.NewWriter(&out)
	enc := sideband64k.NewEncoder(bw)
	enc.SetMaxData(3)

	cw := sideband64k.NewChunkWriter(enc, sideband64k.BandProgress)

	n, err := cw.Write([]byte("abcdefg"))
	if err != nil {
		t.Fatalf("Write: %v", err)
	}

	if n != 7 {
		t.Fatalf("Write n=%d, want 7", n)
	}

	err = enc.Flush()
	if err != nil {
		t.Fatalf("Flush: %v", err)
	}

	if got, want := out.String(), "0008\x02abc0008\x02def0006\x02g"; got != want {
		t.Fatalf("got %q, want %q", got, want)
	}

	out.Reset()

	rn, err := cw.ReadFrom(strings.NewReader("wxyz"))
	if err != nil {
		t.Fatalf("ReadFrom: %v", err)
	}

	if rn != 4 {
		t.Fatalf("ReadFrom n=%d, want 4", rn)
	}

	err = enc.Flush()
	if err != nil {
		t.Fatalf("Flush: %v", err)
	}

	if got, want := out.String(), "0008\x02wxy0006\x02z"; got != want {
		t.Fatalf("got %q, want %q", got, want)
	}
}
