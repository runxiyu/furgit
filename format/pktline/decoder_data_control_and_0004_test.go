package pktline_test

import (
	"strings"
	"testing"
	"codeberg.org/lindenii/furgit/format/pktline"
)

func TestDecoderDataControlAnd0004(t *testing.T) {
	t.Parallel()

	input := "0006a\n0004000100020000"
	dec := pktline.NewDecoder(strings.NewReader(input), pktline.ReadOptions{ChompLF: true})

	f, err := dec.ReadFrame()
	if err != nil {
		t.Fatalf("ReadFrame #1: %v", err)
	}

	if f.Type != pktline.PacketData || string(f.Payload) != "a" {
		t.Fatalf("frame #1 = %#v", f)
	}

	f, err = dec.ReadFrame()
	if err != nil {
		t.Fatalf("ReadFrame #2: %v", err)
	}

	if f.Type != pktline.PacketData || len(f.Payload) != 0 {
		t.Fatalf("frame #2 = %#v, want empty data", f)
	}

	f, err = dec.ReadFrame()
	if err != nil {
		t.Fatalf("ReadFrame #3: %v", err)
	}

	if f.Type != pktline.PacketDelim {
		t.Fatalf("frame #3 type = %v, want PacketDelim", f.Type)
	}

	f, err = dec.ReadFrame()
	if err != nil {
		t.Fatalf("ReadFrame #4: %v", err)
	}

	if f.Type != pktline.PacketResponseEnd {
		t.Fatalf("frame #4 type = %v, want PacketResponseEnd", f.Type)
	}

	f, err = dec.ReadFrame()
	if err != nil {
		t.Fatalf("ReadFrame #5: %v", err)
	}

	if f.Type != pktline.PacketFlush {
		t.Fatalf("frame #5 type = %v, want PacketFlush", f.Type)
	}
}

