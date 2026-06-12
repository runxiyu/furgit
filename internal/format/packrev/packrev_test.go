package packrev_test

import (
	"bytes"
	"encoding/binary"
	"errors"
	"os"
	"testing"

	"lindenii.org/go/furgit/internal/format/packidx"
	"lindenii.org/go/furgit/internal/format/packrev"
	"lindenii.org/go/furgit/object/id"
)

func TestParseGitReverseIndex(t *testing.T) {
	t.Parallel()

	for _, objectFormat := range id.SupportedObjectFormats() {
		t.Run(objectFormat.String(), func(t *testing.T) {
			t.Parallel()

			prefix := makeGitPack(t, objectFormat)

			revData, err := os.ReadFile(prefix + ".rev") //nolint:gosec
			if err != nil {
				t.Fatalf("ReadFile: %v", err)
			}

			rev, err := packrev.Parse(revData, objectFormat)
			if err != nil {
				t.Fatalf("Parse: %v", err)
			}

			idxData, err := os.ReadFile(prefix + ".idx") //nolint:gosec
			if err != nil {
				t.Fatalf("ReadFile: %v", err)
			}

			idx, err := packidx.Parse(idxData, objectFormat.Size())
			if err != nil {
				t.Fatalf("packidx.Parse: %v", err)
			}

			if rev.NumObjects() != idx.NumObjects() {
				t.Fatalf("NumObjects = %d, want %d", rev.NumObjects(), idx.NumObjects())
			}

			packData, err := os.ReadFile(prefix + ".pack") //nolint:gosec
			if err != nil {
				t.Fatalf("ReadFile: %v", err)
			}

			packTrailer := packData[len(packData)-objectFormat.Size():]
			if !bytes.Equal(rev.PackHash(), packTrailer) {
				t.Fatalf("PackHash does not match pack trailer")
			}

			want := packOrderPositions(t, &idx)

			for packOrder, wantPosition := range want {
				position, err := rev.PositionAt(packOrder)
				if err != nil {
					t.Fatalf("PositionAt(%d): %v", packOrder, err)
				}

				if position != int(wantPosition) {
					t.Fatalf("PositionAt(%d) = %d, want %d", packOrder, position, wantPosition)
				}
			}
		})
	}
}

func TestParseMalformed(t *testing.T) {
	t.Parallel()

	for _, objectFormat := range id.SupportedObjectFormats() {
		t.Run(objectFormat.String(), func(t *testing.T) {
			t.Parallel()

			valid := writeSyntheticRev(t, objectFormat, []uint32{2, 0, 1, 3})

			corrupt := func(mutate func(data []byte) []byte) []byte {
				return mutate(bytes.Clone(valid))
			}

			cases := []struct {
				name string
				data []byte
			}{
				{name: "empty", data: []byte{}},
				{name: "truncated", data: corrupt(func(d []byte) []byte { return d[:10] })},
				{
					name: "bad signature",
					data: corrupt(func(d []byte) []byte {
						d[0] ^= 0xff

						return d
					}),
				},
				{
					name: "bad version",
					data: corrupt(func(d []byte) []byte {
						d[7] = 2

						return d
					}),
				},
				{
					name: "hash function mismatch",
					data: corrupt(func(d []byte) []byte {
						d[11] ^= 0xff

						return d
					}),
				},
				{
					name: "position table size not 32-bit multiple",
					data: corrupt(func(d []byte) []byte { return append(d, 0xde, 0xad) }),
				},
			}

			for _, tc := range cases {
				t.Run(tc.name, func(t *testing.T) {
					t.Parallel()

					_, err := packrev.Parse(tc.data, objectFormat)
					if !errors.Is(err, packrev.ErrMalformedReverseIndex) {
						t.Fatalf("Parse error = %v, want ErrMalformedReverseIndex", err)
					}
				})
			}
		})
	}
}

func TestPositionAtMalformed(t *testing.T) {
	t.Parallel()

	for _, objectFormat := range id.SupportedObjectFormats() {
		t.Run(objectFormat.String(), func(t *testing.T) {
			t.Parallel()

			data := writeSyntheticRev(t, objectFormat, []uint32{2, 0, 1, 3})

			// Corrupt the first stored position to one past the object count.
			binary.BigEndian.PutUint32(data[12:], 4)

			rev, err := packrev.Parse(data, objectFormat)
			if err != nil {
				t.Fatalf("Parse: %v", err)
			}

			_, err = rev.PositionAt(0)
			if !errors.Is(err, packrev.ErrMalformedReverseIndex) {
				t.Fatalf("PositionAt error = %v, want ErrMalformedReverseIndex", err)
			}
		})
	}
}
