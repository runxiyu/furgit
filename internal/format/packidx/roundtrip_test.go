package packidx_test

import (
	"bytes"
	"math"
	"testing"

	"lindenii.org/go/furgit/internal/format/packidx"
	"lindenii.org/go/furgit/object/id"
)

func TestWriteRoundTrip(t *testing.T) {
	t.Parallel()

	for _, objectFormat := range id.SupportedObjectFormats() {
		t.Run(objectFormat.String(), func(t *testing.T) {
			t.Parallel()

			hashSize := objectFormat.Size()

			entries := syntheticEntries(20)
			entries[3].Offset = 1 << 33
			entries[11].Offset = math.MaxUint64
			entries[12].Offset = 0x7fffffff
			entries[13].Offset = 0x80000000

			data := writeSyntheticIndex(t, objectFormat, entries)

			idx, err := packidx.Parse(data, hashSize)
			if err != nil {
				t.Fatalf("Parse: %v", err)
			}

			if idx.NumObjects() != len(entries) {
				t.Fatalf("NumObjects = %d, want %d", idx.NumObjects(), len(entries))
			}

			if !bytes.Equal(idx.PackHash(), bytes.Repeat([]byte{0x5a}, hashSize)) {
				t.Fatalf("PackHash mismatch")
			}

			for pos, entry := range entries {
				if !bytes.Equal(idx.OIDAt(pos), entry.OID[:hashSize]) {
					t.Fatalf("OIDAt(%d) mismatch", pos)
				}

				if idx.CRCAt(pos) != entry.CRC32 {
					t.Fatalf("CRCAt(%d) = %x, want %x", pos, idx.CRCAt(pos), entry.CRC32)
				}

				offset, err := idx.OffsetAt(pos)
				if err != nil {
					t.Fatalf("OffsetAt(%d): %v", pos, err)
				}

				if offset != entry.Offset {
					t.Fatalf("OffsetAt(%d) = %d, want %d", pos, offset, entry.Offset)
				}

				offset, found, err := idx.Lookup(entry.OID[:hashSize])
				if err != nil {
					t.Fatalf("Lookup(%d): %v", pos, err)
				}

				if !found || offset != entry.Offset {
					t.Fatalf("Lookup(%d) = %d %v, want %d found", pos, offset, found, entry.Offset)
				}
			}
		})
	}
}

// writeSyntheticIndex writes one index over entries
// with a fixed fake pack hash.
func writeSyntheticIndex(t *testing.T, objectFormat id.ObjectFormat, entries []packidx.Entry) []byte {
	t.Helper()

	packHash := bytes.Repeat([]byte{0x5a}, objectFormat.Size())

	var buf bytes.Buffer

	err := packidx.Write(&buf, objectFormat, entries, packHash)
	if err != nil {
		t.Fatalf("Write: %v", err)
	}

	return buf.Bytes()
}
