package furgit

import "testing"

func TestParseHashValidAndInvalid(t *testing.T) {
	const hex40 = "0123456789abcdef0123456789abcdef01234567"
	id, err := ParseHash(hex40)
	if err != nil {
		t.Fatalf("ParseHash returned error: %v", err)
	}
	if got := id.String(); got != hex40 {
		t.Fatalf("unexpected String result: %q", got)
	}

	if _, err := ParseHash("abcd"); err == nil {
		t.Fatal("expected error for short hash")
	}
	if _, err := ParseHash("zzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzz"); err == nil {
		t.Fatal("expected error for non-hex input")
	}
}

func TestHashBytesCopiesUnderlyingData(t *testing.T) {
	var id Hash
	for i := range id {
		id[i] = byte(i)
	}
	orig := id.Bytes()
	orig[0] ^= 0xff
	if id[0] == orig[0] {
		t.Fatal("Bytes should return a copy")
	}
}
