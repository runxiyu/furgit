package packidx_test

import (
	"bytes"
	"encoding/binary"
	"errors"
	"testing"

	"lindenii.org/go/furgit/internal/format/packidx"
	"lindenii.org/go/furgit/object/id"
)

func TestLookupGitIndex(t *testing.T) {
	t.Parallel()

	for _, objectFormat := range id.SupportedObjectFormats() {
		t.Run(objectFormat.String(), func(t *testing.T) {
			t.Parallel()

			repo, prefix, oids := makeGitPack(t, objectFormat)
			_, idx := parseGitIdxFile(t, prefix, objectFormat)

			wantOffsets := gitPackOffsets(t, repo, prefix+".idx", objectFormat)
			if len(wantOffsets) != len(oids) {
				t.Fatalf("verify-pack offsets = %d entries, want %d", len(wantOffsets), len(oids))
			}

			for _, oid := range oids {
				offset, found, err := idx.Lookup(oid.RawBytes())
				if err != nil {
					t.Fatalf("Lookup(%s): %v", oid, err)
				}

				if !found {
					t.Fatalf("Lookup(%s) not found", oid)
				}

				if offset != wantOffsets[oid.String()] {
					t.Fatalf("Lookup(%s) = %d, want %d", oid, offset, wantOffsets[oid.String()])
				}
			}
		})
	}
}

func TestLookupMissing(t *testing.T) {
	t.Parallel()

	for _, objectFormat := range id.SupportedObjectFormats() {
		t.Run(objectFormat.String(), func(t *testing.T) {
			t.Parallel()

			_, prefix, oids := makeGitPack(t, objectFormat)
			_, idx := parseGitIdxFile(t, prefix, objectFormat)

			missing := bytes.Clone(oids[0].RawBytes())
			missing[len(missing)-1] ^= 0xff

			_, found, err := idx.Lookup(missing)
			if err != nil {
				t.Fatalf("Lookup: %v", err)
			}

			if found {
				t.Fatalf("Lookup of mutated oid unexpectedly found")
			}
		})
	}
}

func TestOffsetAtMalformedLargeReference(t *testing.T) {
	t.Parallel()

	for _, objectFormat := range id.SupportedObjectFormats() {
		t.Run(objectFormat.String(), func(t *testing.T) {
			t.Parallel()

			hashSize := objectFormat.Size()

			entries := syntheticEntries(3)
			data := writeSyntheticIndex(t, objectFormat, entries)

			// Mark the first 32-bit offset entry as a 64-bit reference;
			// the index has no 64-bit offset table.
			off32Off := 8 + 256*4 + len(entries)*hashSize + len(entries)*4
			binary.BigEndian.PutUint32(data[off32Off:], 0x80000000)

			idx, err := packidx.Parse(data, hashSize)
			if err != nil {
				t.Fatalf("Parse: %v", err)
			}

			_, err = idx.OffsetAt(0)
			if !errors.Is(err, packidx.ErrMalformedPackIndex) {
				t.Fatalf("OffsetAt error = %v, want ErrMalformedPackIndex", err)
			}
		})
	}
}
