package id_test

import (
	"bytes"
	"strings"
	"testing"

	"lindenii.org/go/furgit/object/id"
)

func TestFromBytesErrors(t *testing.T) {
	t.Parallel()

	_, err := id.ObjectFormatUnknown.FromBytes([]byte{1, 2})
	if err == nil {
		t.Fatalf("expected FromBytes unknown object format error")
	}

	for _, objectFormat := range id.SupportedObjectFormats() {
		_, err = objectFormat.FromBytes([]byte{1, 2})
		if err == nil {
			t.Fatalf("expected FromBytes wrong size error")
		}
	}
}

func TestBytesReturnsCopy(t *testing.T) {
	t.Parallel()

	for _, objectFormat := range id.SupportedObjectFormats() {
		id, err := objectFormat.FromString(strings.Repeat("01", objectFormat.Size()))
		if err != nil {
			t.Fatalf("FromString failed: %v", err)
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

func TestFromStringErrors(t *testing.T) {
	t.Parallel()

	t.Run("unknown object format", func(t *testing.T) {
		t.Parallel()

		_, err := id.ObjectFormatUnknown.FromString("00")
		if err == nil {
			t.Fatalf("expected FromString error")
		}
	})
	// TODO: This may need to be revisited when hash-function-transition is implemented.

	for _, objectFormat := range id.SupportedObjectFormats() {
		t.Run(objectFormat.String(), func(t *testing.T) {
			t.Parallel()

			_, err := objectFormat.FromString(strings.Repeat("0", objectFormat.HexLen()-1))
			if err == nil {
				t.Fatalf("expected FromString odd-len error")
			}

			_, err = objectFormat.FromString(strings.Repeat("0", objectFormat.HexLen()-2))
			if err == nil {
				t.Fatalf("expected FromString wrong-len error")
			}

			_, err = objectFormat.FromString("z" + strings.Repeat("0", objectFormat.HexLen()-1))
			if err == nil {
				t.Fatalf("expected FromString invalid-hex error")
			}
		})
	}
}

func TestFromStringRoundtrip(t *testing.T) {
	t.Parallel()

	for _, objectFormat := range id.SupportedObjectFormats() {
		t.Run(objectFormat.String(), func(t *testing.T) {
			t.Parallel()

			hex := strings.Repeat("01", objectFormat.Size())

			i, err := objectFormat.FromString(hex)
			if err != nil {
				t.Fatalf("FromString failed: %v", err)
			}

			if got := i.String(); got != hex {
				t.Fatalf("String() = %q, want %q", got, hex)
			}

			if got := i.ObjectFormat().Size(); got != objectFormat.Size() {
				t.Fatalf("Size() = %d, want %d", got, objectFormat.Size())
			}

			raw := i.Bytes()
			if len(raw) != objectFormat.Size() {
				t.Fatalf("Bytes len = %d, want %d", len(raw), objectFormat.Size())
			}

			id2, err := objectFormat.FromBytes(raw)
			if err != nil {
				t.Fatalf("FromBytes failed: %v", err)
			}

			if id2.String() != hex {
				t.Fatalf("FromBytes roundtrip = %q, want %q", id2.String(), hex)
			}
		})
	}
}
