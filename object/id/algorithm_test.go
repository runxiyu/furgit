package id_test

import (
	"bytes"
	"strings"
	"testing"

	"codeberg.org/lindenii/furgit/object/id"
)

func TestAlgorithmEmptyTree(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		algo id.Algorithm
		want string
	}{
		{
			name: "sha1",
			algo: id.AlgorithmSHA1,
			want: "4b825dc642cb6eb9a060e54bf8d69288fbee4904",
		},
		{
			name: "sha256",
			algo: id.AlgorithmSHA256,
			want: "6ef19b41225c5369f1c104d45d8d85efa9b057b53b14b4b9b939dd74decc5321",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := tt.algo.EmptyTree()

			if got.String() != tt.want {
				t.Fatalf("EmptyTree() = %q, want %q", got.String(), tt.want)
			}
		})
	}
}

func TestParseAlgorithm(t *testing.T) {
	t.Parallel()

	algo, ok := id.ParseAlgorithm("sha1")
	if !ok || algo != id.AlgorithmSHA1 {
		t.Fatalf("ParseAlgorithm(sha1) = (%v,%v)", algo, ok)
	}

	algo, ok = id.ParseAlgorithm("sha256")
	if !ok || algo != id.AlgorithmSHA256 {
		t.Fatalf("ParseAlgorithm(sha256) = (%v,%v)", algo, ok)
	}

	if _, ok := id.ParseAlgorithm("md5"); ok {
		t.Fatalf("ParseAlgorithm(md5) should fail")
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

func TestAlgorithmSum(t *testing.T) {
	t.Parallel()

	id1 := id.AlgorithmSHA1.Sum([]byte("hello"))
	if id1.Algorithm() != id.AlgorithmSHA1 || id1.Algorithm().Size() != id.AlgorithmSHA1.Size() {
		t.Fatalf("sha1 sum produced invalid object id")
	}

	id2 := id.AlgorithmSHA256.Sum([]byte("hello"))
	if id2.Algorithm() != id.AlgorithmSHA256 || id2.Algorithm().Size() != id.AlgorithmSHA256.Size() {
		t.Fatalf("sha256 sum produced invalid object id")
	}

	if id1.String() == id2.String() {
		t.Fatalf("sha1 and sha256 should differ")
	}
}

func TestUnknownAlgorithmEmptyTree(t *testing.T) {
	t.Parallel()

	got := id.AlgorithmUnknown.EmptyTree()
	if got != (id.ObjectID{}) {
		t.Fatalf("EmptyTree() for unknown algorithm = %#v, want zero value", got)
	}
}
