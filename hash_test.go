package furgit

import (
	"strings"
	"testing"
)

func TestParseHashValidAndInvalid(t *testing.T) {
	pattern := "0123456789abcdef"
	repeats := (testHashSize*2 + len(pattern) - 1) / len(pattern)
	hexStr := strings.Repeat(pattern, repeats)[:testHashSize*2]
	
	repo := &Repository{HashSize: testHashSize}
	id, err := repo.ParseHash(hexStr)
	if err != nil {
		t.Fatalf("ParseHash returned error: %v", err)
	}

	if got := id.String(); got != hexStr {
		t.Fatalf("unexpected String result: %q", got)
	}

	if _, err := repo.ParseHash("abcd"); err == nil {
		t.Fatal("expected error for short hash")
	}

	badHex := strings.Repeat("z", testHashSize*2)
	if _, err := repo.ParseHash(badHex); err == nil {
		t.Fatal("expected error for non-hex input")
	}
}

func TestHashBytesCopiesUnderlyingData(t *testing.T) {
	var id Hash
	for i := range id.data {
		id.data[i] = byte(i)
	}
	id.size = testHashSize
	orig := id.Bytes()
	orig[0] ^= 0xff
	if id.data[0] == orig[0] {
		t.Fatal("Bytes should return a copy")
	}
}
