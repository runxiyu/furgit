package oid

import (
	"bytes"
	"testing"
)

func TestParseAlgorithm(t *testing.T) {
	t.Parallel()

	algo, ok := ParseAlgorithm("sha1")
	if !ok || algo != AlgorithmSHA1 {
		t.Fatalf("ParseAlgorithm(sha1) = (%v,%v)", algo, ok)
	}

	algo, ok = ParseAlgorithm("sha256")
	if !ok || algo != AlgorithmSHA256 {
		t.Fatalf("ParseAlgorithm(sha256) = (%v,%v)", algo, ok)
	}

	if _, ok := ParseAlgorithm("md5"); ok {
		t.Fatalf("ParseAlgorithm(md5) should fail")
	}
}

func TestParseHexRoundtrip(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		algo Algorithm
		hex  string
	}{
		{
			name: "sha1",
			algo: AlgorithmSHA1,
			hex:  "0123456789abcdef0123456789abcdef01234567",
		},
		{
			name: "sha256",
			algo: AlgorithmSHA256,
			hex:  "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			id, err := ParseHex(tt.algo, tt.hex)
			if err != nil {
				t.Fatalf("ParseHex failed: %v", err)
			}
			if got := id.String(); got != tt.hex {
				t.Fatalf("String() = %q, want %q", got, tt.hex)
			}
			if got := id.Size(); got != tt.algo.Size() {
				t.Fatalf("Size() = %d, want %d", got, tt.algo.Size())
			}

			raw := id.Bytes()
			if len(raw) != tt.algo.Size() {
				t.Fatalf("Bytes len = %d, want %d", len(raw), tt.algo.Size())
			}

			id2, err := FromBytes(tt.algo, raw)
			if err != nil {
				t.Fatalf("FromBytes failed: %v", err)
			}
			if id2.String() != tt.hex {
				t.Fatalf("FromBytes roundtrip = %q, want %q", id2.String(), tt.hex)
			}
		})
	}
}

func TestParseHexErrors(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		algo Algorithm
		hex  string
	}{
		{"unknown algo", AlgorithmUnknown, "00"},
		{"odd len", AlgorithmSHA1, "0"},
		{"wrong len", AlgorithmSHA1, "0123"},
		{"invalid hex", AlgorithmSHA1, "zz23456789abcdef0123456789abcdef01234567"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if _, err := ParseHex(tt.algo, tt.hex); err == nil {
				t.Fatalf("expected ParseHex error")
			}
		})
	}
}

func TestFromBytesErrors(t *testing.T) {
	t.Parallel()

	if _, err := FromBytes(AlgorithmUnknown, []byte{1, 2}); err == nil {
		t.Fatalf("expected FromBytes unknown algo error")
	}
	if _, err := FromBytes(AlgorithmSHA1, []byte{1, 2}); err == nil {
		t.Fatalf("expected FromBytes wrong size error")
	}
}

func TestBytesReturnsCopy(t *testing.T) {
	t.Parallel()

	id, err := ParseHex(AlgorithmSHA1, "0123456789abcdef0123456789abcdef01234567")
	if err != nil {
		t.Fatalf("ParseHex failed: %v", err)
	}

	b1 := id.Bytes()
	b2 := id.Bytes()
	if !bytes.Equal(b1, b2) {
		t.Fatalf("Bytes mismatch")
	}
	b1[0] ^= 0xff
	if bytes.Equal(b1, b2) {
		t.Fatalf("Bytes should return independent copies")
	}
}

func TestAlgorithmSum(t *testing.T) {
	t.Parallel()

	id1 := AlgorithmSHA1.Sum([]byte("hello"))
	if id1.Algorithm() != AlgorithmSHA1 || id1.Size() != AlgorithmSHA1.Size() {
		t.Fatalf("sha1 sum produced invalid object id")
	}

	id2 := AlgorithmSHA256.Sum([]byte("hello"))
	if id2.Algorithm() != AlgorithmSHA256 || id2.Size() != AlgorithmSHA256.Size() {
		t.Fatalf("sha256 sum produced invalid object id")
	}

	if id1.String() == id2.String() {
		t.Fatalf("sha1 and sha256 should differ")
	}
}
