package furgit

import (
	"strings"
	"testing"
)

func TestParseHashValidAndInvalid(t *testing.T) {
	pattern := "0123456789abcdef"
	repeats := (testHashSize*2 + len(pattern) - 1) / len(pattern)
	hexStr := strings.Repeat(pattern, repeats)[:testHashSize*2]

	id, err := ParseHash[testHashType](hexStr)
	if err != nil {
		t.Fatalf("ParseHash returned error: %v", err)
	}

	if got := id.String(); got != hexStr {
		t.Fatalf("unexpected String result: %q", got)
	}

	if _, err := ParseHash[testHashType]("abcd"); err == nil {
		t.Fatal("expected error for short hash")
	}

	badHex := strings.Repeat("z", testHashSize*2)
	if _, err := ParseHash[testHashType](badHex); err == nil {
		t.Fatal("expected error for non-hex input")
	}
}

func TestHashTypeCopiesUnderlyingData(t *testing.T) {
	var id TestHash
	idSlice := id.Slice()
	for i := range idSlice {
		idSlice[i] = byte(i)
	}
	orig := id.Bytes()
	orig[0] ^= 0xff
	if idSlice[0] == orig[0] {
		t.Fatal("Bytes should return a copy")
	}
}
