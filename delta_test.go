package furgit

import (
	"bytes"
	"testing"

	"codeberg.org/lindenii/furgit/internal/bufpool"
)

func TestDeltaEncodeApplyRoundtrip(t *testing.T) {
	base := []byte("Hello, world!")
	target := []byte("Hello, brave new world!")

	instr := []deltaInstruction{
		{copy: true, offset: 0, length: 7},
		{copy: false, data: []byte("brave new "), length: len("brave new ")},
		{copy: true, offset: 7, length: 6},
	}
	delta, err := deltaEncode(len(base), len(target), instr)
	if err != nil {
		t.Fatalf("deltaEncode error: %v", err)
	}

	out, err := packDeltaApply(bufpool.FromOwned(base), bufpool.FromOwned(delta))
	if err != nil {
		t.Fatalf("packDeltaApply error: %v", err)
	}
	defer out.Release()
	if !bytes.Equal(out.Bytes(), target) {
		t.Fatalf("delta apply mismatch: got %q, want %q", out.Bytes(), target)
	}
}

func TestDeltifyRoundtrip(t *testing.T) {
	base := []byte("The quick brown fox jumps over the lazy dog.")
	target := []byte("The quick red fox jumps over the very lazy dog.")

	instr, err := deltify(base, target, 0)
	if err != nil {
		t.Fatalf("deltify error: %v", err)
	}
	delta, err := deltaEncode(len(base), len(target), instr)
	if err != nil {
		t.Fatalf("deltaEncode error: %v", err)
	}

	out, err := packDeltaApply(bufpool.FromOwned(base), bufpool.FromOwned(delta))
	if err != nil {
		t.Fatalf("packDeltaApply error: %v", err)
	}
	defer out.Release()
	if !bytes.Equal(out.Bytes(), target) {
		t.Fatalf("delta apply mismatch: got %q, want %q", out.Bytes(), target)
	}
}
