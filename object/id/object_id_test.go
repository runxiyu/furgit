package id_test

import (
	"bytes"
	"strings"
	"testing"

	"codeberg.org/lindenii/furgit/object/id"
)

func TestFromBytesErrors(t *testing.T) {
	t.Parallel()

	_, err := id.FromBytes(id.AlgorithmUnknown, []byte{1, 2})
	if err == nil {
		t.Fatalf("expected FromBytes unknown algo error")
	}

	for _, algo := range id.SupportedAlgorithms() {
		_, err = id.FromBytes(algo, []byte{1, 2})
		if err == nil {
			t.Fatalf("expected FromBytes wrong size error")
		}
	}
}

func TestBytesReturnsCopy(t *testing.T) {
	t.Parallel()

	for _, algo := range id.SupportedAlgorithms() {
		id, err := id.FromHex(algo, strings.Repeat("01", algo.Size()))
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

func TestFromHexErrors(t *testing.T) {
	t.Parallel()

	t.Run("unknown algo", func(t *testing.T) {
		t.Parallel()

		_, err := id.FromHex(id.AlgorithmUnknown, "00")
		if err == nil {
			t.Fatalf("expected FromHex error")
		}
	})
	// TODO: This may need to be revisited when hash-function-transition is implemented.

	for _, algo := range id.SupportedAlgorithms() {
		t.Run(algo.String(), func(t *testing.T) {
			t.Parallel()

			_, err := id.FromHex(algo, strings.Repeat("0", algo.HexLen()-1))
			if err == nil {
				t.Fatalf("expected FromHex odd-len error")
			}

			_, err = id.FromHex(algo, strings.Repeat("0", algo.HexLen()-2))
			if err == nil {
				t.Fatalf("expected FromHex wrong-len error")
			}

			_, err = id.FromHex(algo, "z"+strings.Repeat("0", algo.HexLen()-1))
			if err == nil {
				t.Fatalf("expected FromHex invalid-hex error")
			}
		})
	}
}

func TestFromHexRoundtrip(t *testing.T) {
	t.Parallel()

	for _, algo := range id.SupportedAlgorithms() {
		t.Run(algo.String(), func(t *testing.T) {
			t.Parallel()

			hex := strings.Repeat("01", algo.Size())

			i, err := id.FromHex(algo, hex)
			if err != nil {
				t.Fatalf("FromHex failed: %v", err)
			}

			if got := i.String(); got != hex {
				t.Fatalf("String() = %q, want %q", got, hex)
			}

			if got := i.Algorithm().Size(); got != algo.Size() {
				t.Fatalf("Size() = %d, want %d", got, algo.Size())
			}

			raw := i.Bytes()
			if len(raw) != algo.Size() {
				t.Fatalf("Bytes len = %d, want %d", len(raw), algo.Size())
			}

			id2, err := id.FromBytes(algo, raw)
			if err != nil {
				t.Fatalf("FromBytes failed: %v", err)
			}

			if id2.String() != hex {
				t.Fatalf("FromBytes roundtrip = %q, want %q", id2.String(), hex)
			}
		})
	}
}
