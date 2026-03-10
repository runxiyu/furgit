package pktline_test

import (
	"testing"

	"codeberg.org/lindenii/furgit/protocol/pktline"
)

func TestAppendHelpers(t *testing.T) {
	t.Parallel()

	out, err := pktline.AppendData(nil, []byte("ok"))
	if err != nil {
		t.Fatalf("AppendData: %v", err)
	}

	out = pktline.AppendFlushPkt(out)
	out = pktline.AppendDelimPkt(out)
	out = pktline.AppendResponseEndPkt(out)

	if got, want := string(out), "0006ok000000010002"; got != want {
		t.Fatalf("got %q, want %q", got, want)
	}
}
