package packidx_test

import (
	"bytes"
	"errors"
	"testing"

	"lindenii.org/go/furgit/internal/format/packidx"
	"lindenii.org/go/furgit/object/id"
)

// syntheticEntries builds n distinct entries sorted by object ID,
// spread over the fanout table.
func syntheticEntries(n int) []packidx.Entry {
	entries := make([]packidx.Entry, n)
	for i := range entries {
		entries[i].OID[0] = byte(i * 7)
		entries[i].OID[1] = byte(i + 1)
		entries[i].Offset = uint64(i+1) * 100
		entries[i].CRC32 = uint32(i+1) * 0x01010101
	}

	return entries
}

func TestWriteMatchesGit(t *testing.T) {
	t.Parallel()

	for _, objectFormat := range id.SupportedObjectFormats() {
		t.Run(objectFormat.String(), func(t *testing.T) {
			t.Parallel()

			_, prefix, _ := makeGitPack(t, objectFormat)
			gitData, idx := parseGitIdxFile(t, prefix, objectFormat)

			entries := make([]packidx.Entry, idx.NumObjects())
			for pos := range entries {
				copy(entries[pos].OID[:], idx.OIDAt(pos))
				entries[pos].CRC32 = idx.CRCAt(pos)

				offset, err := idx.OffsetAt(pos)
				if err != nil {
					t.Fatalf("OffsetAt(%d): %v", pos, err)
				}

				entries[pos].Offset = offset
			}

			var buf bytes.Buffer

			err := packidx.Write(&buf, objectFormat, entries, idx.PackHash())
			if err != nil {
				t.Fatalf("Write: %v", err)
			}

			if !bytes.Equal(buf.Bytes(), gitData) {
				t.Fatalf("Write output differs from git's index (%d vs %d bytes)", buf.Len(), len(gitData))
			}
		})
	}
}

func TestWriteInvalidEntries(t *testing.T) {
	t.Parallel()

	unsorted := syntheticEntries(3)
	unsorted[0], unsorted[2] = unsorted[2], unsorted[0]

	duplicated := syntheticEntries(3)
	duplicated[1] = duplicated[0]

	cases := []struct {
		name    string
		entries []packidx.Entry
	}{
		{name: "unsorted", entries: unsorted},
		{name: "duplicated", entries: duplicated},
	}

	for _, objectFormat := range id.SupportedObjectFormats() {
		t.Run(objectFormat.String(), func(t *testing.T) {
			t.Parallel()

			packHash := bytes.Repeat([]byte{0x5a}, objectFormat.Size())

			for _, tc := range cases {
				t.Run(tc.name, func(t *testing.T) {
					t.Parallel()

					err := packidx.Write(&bytes.Buffer{}, objectFormat, tc.entries, packHash)
					if !errors.Is(err, packidx.ErrInvalidEntries) {
						t.Fatalf("Write error = %v, want ErrInvalidEntries", err)
					}
				})
			}
		})
	}
}

func TestWriteBadPackHashPanics(t *testing.T) {
	t.Parallel()

	objectFormat := id.ObjectFormatSHA256

	defer func() {
		if recover() == nil {
			t.Fatalf("Write with short pack hash: expected panic")
		}
	}()

	_ = packidx.Write(&bytes.Buffer{}, objectFormat, nil, []byte{0x01})
}
