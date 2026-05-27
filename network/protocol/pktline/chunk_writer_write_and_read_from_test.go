package pktline_test

import (
	"bufio"
	"bytes"
	"strings"
	"testing"

	"lindenii.org/go/furgit/network/protocol/pktline"
)

func TestChunkWriterWriteAndReadFrom(t *testing.T) {
	t.Parallel()

	var out bytes.Buffer

	bw := bufio.NewWriter(&out)

	enc := pktline.NewEncoder(bw)
	enc.SetMaxData(3)
	cw := pktline.NewChunkWriter(enc)

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

	if got, want := out.String(), "0007abc0007def0005g"; got != want {
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

	if got, want := out.String(), "0007wxy0005z"; got != want {
		t.Fatalf("got %q, want %q", got, want)
	}
}
