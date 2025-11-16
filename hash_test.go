package furgit

import (
	"strings"
	"testing"
)

func TestParseHashValidAndInvalid(t *testing.T) {
	pattern := "0123456789abcdef"
	repeats := (testHashSize*2 + len(pattern) - 1) / len(pattern)
	hexStr := strings.Repeat(pattern, repeats)[:testHashSize*2]

	id, err := ParseHashWithSize(hexStr, testHashSize)
	if err != nil {
		t.Fatalf("ParseHash returned error: %v", err)
	}

	if got := id.StringWithSize(testHashSize); got != hexStr {
		t.Fatalf("unexpected String result: %q", got)
	}

	if _, err := ParseHashWithSize("abcd", testHashSize); err == nil {
		t.Fatal("expected error for short hash")
	}

	badHex := strings.Repeat("z", testHashSize*2)
	if _, err := ParseHashWithSize(badHex, testHashSize); err == nil {
		t.Fatal("expected error for non-hex input")
	}
}

func TestHashBytesCopiesUnderlyingData(t *testing.T) {
	var id Hash
	for i := range id {
		id[i] = byte(i)
	}
	orig := id.BytesWithSize(testHashSize)
	orig[0] ^= 0xff
	if id[0] == orig[0] {
		t.Fatal("Bytes should return a copy")
	}
}
