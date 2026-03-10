package sideband64k_test

import (
	"strings"
	"testing"

	"codeberg.org/lindenii/furgit/protocol/sideband64k"
)

func TestDecoderDataControlAndKeepalive(t *testing.T) {
	t.Parallel()

	input := "0007\x01a\n0005\x010007\x02p\n0007\x03e\n000100020000"
	dec := sideband64k.NewDecoder(strings.NewReader(input), sideband64k.ReadOptions{ChompLF: true})

	f, err := dec.ReadFrame()
	if err != nil {
		t.Fatalf("ReadFrame #1: %v", err)
	}

	if f.Type != sideband64k.FrameData || string(f.Payload) != "a" {
		t.Fatalf("frame #1 = %#v", f)
	}

	f, err = dec.ReadFrame()
	if err != nil {
		t.Fatalf("ReadFrame #2: %v", err)
	}

	if f.Type != sideband64k.FrameData || len(f.Payload) != 0 {
		t.Fatalf("frame #2 = %#v, want empty data", f)
	}

	f, err = dec.ReadFrame()
	if err != nil {
		t.Fatalf("ReadFrame #3: %v", err)
	}

	if f.Type != sideband64k.FrameProgress || string(f.Payload) != "p\n" {
		t.Fatalf("frame #3 = %#v", f)
	}

	f, err = dec.ReadFrame()
	if err != nil {
		t.Fatalf("ReadFrame #4: %v", err)
	}

	if f.Type != sideband64k.FrameError || string(f.Payload) != "e\n" {
		t.Fatalf("frame #4 = %#v", f)
	}

	f, err = dec.ReadFrame()
	if err != nil {
		t.Fatalf("ReadFrame #5: %v", err)
	}

	if f.Type != sideband64k.FrameDelim {
		t.Fatalf("frame #5 type = %v, want FrameDelim", f.Type)
	}

	f, err = dec.ReadFrame()
	if err != nil {
		t.Fatalf("ReadFrame #6: %v", err)
	}

	if f.Type != sideband64k.FrameResponseEnd {
		t.Fatalf("frame #6 type = %v, want FrameResponseEnd", f.Type)
	}

	f, err = dec.ReadFrame()
	if err != nil {
		t.Fatalf("ReadFrame #7: %v", err)
	}

	if f.Type != sideband64k.FrameFlush {
		t.Fatalf("frame #7 type = %v, want FrameFlush", f.Type)
	}
}
