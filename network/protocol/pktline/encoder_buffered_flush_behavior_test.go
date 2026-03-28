package pktline_test

import (
	"bufio"
	"bytes"
	"testing"

	"codeberg.org/lindenii/furgit/network/protocol/pktline"
)

func TestEncoderBufferedFlushBehavior(t *testing.T) {
	t.Parallel()

	var out bytes.Buffer

	bw := bufio.NewWriter(&out)
	enc := pktline.NewEncoder(bw)

	err := enc.WriteData([]byte("hello"))
	if err != nil {
		t.Fatalf("WriteData: %v", err)
	}

	err = enc.WriteFlushPacket()
	if err != nil {
		t.Fatalf("WriteFlushPacket: %v", err)
	}

	if out.Len() != 0 {
		t.Fatalf("WriteFlushPacket should not flush I/O, got %q", out.String())
	}

	err = enc.Flush()
	if err != nil {
		t.Fatalf("Flush: %v", err)
	}

	if got, want := out.String(), "0009hello0000"; got != want {
		t.Fatalf("got %q, want %q", got, want)
	}

	out.Reset()
	bw = bufio.NewWriter(&out)
	enc = pktline.NewEncoder(bw)

	err = enc.WriteData([]byte("ok"))
	if err != nil {
		t.Fatalf("WriteData: %v", err)
	}

	err = enc.WriteFlushPacket()
	if err != nil {
		t.Fatalf("WriteFlushPacket: %v", err)
	}

	if out.Len() != 0 {
		t.Fatalf("WriteFlushPacket should not flush I/O, got %q", out.String())
	}

	err = enc.Flush()
	if err != nil {
		t.Fatalf("Flush: %v", err)
	}

	if got, want := out.String(), "0006ok0000"; got != want {
		t.Fatalf("got %q, want %q", got, want)
	}

	out.Reset()
	bw = bufio.NewWriter(&out)
	enc = pktline.NewEncoder(bw)

	err = enc.WriteData([]byte("yo"))
	if err != nil {
		t.Fatalf("WriteData: %v", err)
	}

	err = enc.WriteFlushPacketAndFlush()
	if err != nil {
		t.Fatalf("WriteFlushPacketAndFlush: %v", err)
	}

	if got, want := out.String(), "0006yo0000"; got != want {
		t.Fatalf("got %q, want %q", got, want)
	}
}
