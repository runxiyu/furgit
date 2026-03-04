package objectid_test

import (
	"bytes"
	"strings"
	"testing"

	"codeberg.org/lindenii/furgit/objectid"
)

func TestParseAlgorithm(t *testing.T) {
	t.Parallel()

	algo, ok := objectid.ParseAlgorithm("sha1")
	if !ok || algo != objectid.AlgorithmSHA1 {
		t.Fatalf("ParseAlgorithm(sha1) = (%v,%v)", algo, ok)
	}

	algo, ok = objectid.ParseAlgorithm("sha256")
	if !ok || algo != objectid.AlgorithmSHA256 {
		t.Fatalf("ParseAlgorithm(sha256) = (%v,%v)", algo, ok)
	}

	if _, ok := objectid.ParseAlgorithm("md5"); ok {
		t.Fatalf("ParseAlgorithm(md5) should fail")
	}
}

func TestParseHexRoundtrip(t *testing.T) {
	t.Parallel()

	for _, algo := range objectid.SupportedAlgorithms() {
		t.Run(algo.String(), func(t *testing.T) {
			t.Parallel()

			hex := strings.Repeat("01", algo.Size())

			id, err := objectid.ParseHex(algo, hex)
			if err != nil {
				t.Fatalf("ParseHex failed: %v", err)
			}

			if got := id.String(); got != hex {
				t.Fatalf("String() = %q, want %q", got, hex)
			}

			if got := id.Size(); got != algo.Size() {
				t.Fatalf("Size() = %d, want %d", got, algo.Size())
			}

			raw := id.Bytes()
			if len(raw) != algo.Size() {
				t.Fatalf("Bytes len = %d, want %d", len(raw), algo.Size())
			}

			id2, err := objectid.FromBytes(algo, raw)
			if err != nil {
				t.Fatalf("FromBytes failed: %v", err)
			}

			if id2.String() != hex {
				t.Fatalf("FromBytes roundtrip = %q, want %q", id2.String(), hex)
			}
		})
	}
}

func TestParseHexErrors(t *testing.T) {
	t.Parallel()

	t.Run("unknown algo", func(t *testing.T) {
		t.Parallel()

		_, err := objectid.ParseHex(objectid.AlgorithmUnknown, "00")
		if err == nil {
			t.Fatalf("expected ParseHex error")
		}
	})

	for _, algo := range objectid.SupportedAlgorithms() {
		t.Run(algo.String(), func(t *testing.T) {
			t.Parallel()

			_, err := objectid.ParseHex(algo, strings.Repeat("0", algo.HexLen()-1))
			if err == nil {
				t.Fatalf("expected ParseHex odd-len error")
			}

			_, err = objectid.ParseHex(algo, strings.Repeat("0", algo.HexLen()-2))
			if err == nil {
				t.Fatalf("expected ParseHex wrong-len error")
			}

			_, err = objectid.ParseHex(algo, "z"+strings.Repeat("0", algo.HexLen()-1))
			if err == nil {
				t.Fatalf("expected ParseHex invalid-hex error")
			}
		})
	}
}

func TestFromBytesErrors(t *testing.T) {
	t.Parallel()

	_, err := objectid.FromBytes(objectid.AlgorithmUnknown, []byte{1, 2})
	if err == nil {
		t.Fatalf("expected FromBytes unknown algo error")
	}

	for _, algo := range objectid.SupportedAlgorithms() {
		_, err = objectid.FromBytes(algo, []byte{1, 2})
		if err == nil {
			t.Fatalf("expected FromBytes wrong size error")
		}
	}
}

func TestBytesReturnsCopy(t *testing.T) {
	t.Parallel()

	for _, algo := range objectid.SupportedAlgorithms() {
		id, err := objectid.ParseHex(algo, strings.Repeat("01", algo.Size()))
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
}

func TestRawBytesAliasesStorage(t *testing.T) {
	t.Parallel()

	for _, algo := range objectid.SupportedAlgorithms() {
		id, err := objectid.ParseHex(algo, strings.Repeat("01", algo.Size()))
		if err != nil {
			t.Fatalf("ParseHex failed: %v", err)
		}

		b := id.RawBytes()
		if len(b) != id.Size() {
			t.Fatalf("RawBytes len = %d, want %d", len(b), id.Size())
		}

		if cap(b) != len(b) {
			t.Fatalf("RawBytes cap = %d, want %d", cap(b), len(b))
		}

		orig := id.String()
		b[0] ^= 0xff

		if id.String() == orig {
			t.Fatalf("RawBytes should alias object ID storage")
		}
	}
}

func TestAlgorithmSum(t *testing.T) {
	t.Parallel()

	id1 := objectid.AlgorithmSHA1.Sum([]byte("hello"))
	if id1.Algorithm() != objectid.AlgorithmSHA1 || id1.Size() != objectid.AlgorithmSHA1.Size() {
		t.Fatalf("sha1 sum produced invalid object id")
	}

	id2 := objectid.AlgorithmSHA256.Sum([]byte("hello"))
	if id2.Algorithm() != objectid.AlgorithmSHA256 || id2.Size() != objectid.AlgorithmSHA256.Size() {
		t.Fatalf("sha256 sum produced invalid object id")
	}

	if id1.String() == id2.String() {
		t.Fatalf("sha1 and sha256 should differ")
	}
}
