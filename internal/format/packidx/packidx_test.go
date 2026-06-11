package packidx_test

import (
	"bytes"
	"encoding/binary"
	"errors"
	"os"
	"testing"

	"lindenii.org/go/furgit/internal/format/packidx"
	"lindenii.org/go/furgit/object/id"
)

func TestParseGitIndex(t *testing.T) {
	t.Parallel()

	for _, objectFormat := range id.SupportedObjectFormats() {
		t.Run(objectFormat.String(), func(t *testing.T) {
			t.Parallel()

			_, prefix, oids := makeGitPack(t, objectFormat)
			_, idx := parseGitIdxFile(t, prefix, objectFormat)

			if idx.NumObjects() != len(oids) {
				t.Fatalf("NumObjects = %d, want %d", idx.NumObjects(), len(oids))
			}

			for pos := 1; pos < idx.NumObjects(); pos++ {
				if bytes.Compare(idx.OIDAt(pos-1), idx.OIDAt(pos)) >= 0 {
					t.Fatalf("OIDAt(%d) not sorted after predecessor", pos)
				}
			}

			packData, err := os.ReadFile(prefix + ".pack") //nolint:gosec
			if err != nil {
				t.Fatalf("ReadFile: %v", err)
			}

			packTrailer := packData[len(packData)-objectFormat.Size():]
			if !bytes.Equal(idx.PackHash(), packTrailer) {
				t.Fatalf("PackHash does not match pack trailer")
			}
		})
	}
}

func TestParseMalformed(t *testing.T) {
	t.Parallel()

	for _, objectFormat := range id.SupportedObjectFormats() {
		t.Run(objectFormat.String(), func(t *testing.T) {
			t.Parallel()

			hashSize := objectFormat.Size()

			valid := writeSyntheticIndex(t, objectFormat, syntheticEntries(8))

			corrupt := func(mutate func(data []byte) []byte) []byte {
				return mutate(bytes.Clone(valid))
			}

			cases := []struct {
				name string
				data []byte
			}{
				{name: "empty", data: []byte{}},
				{name: "truncated", data: corrupt(func(d []byte) []byte { return d[:20] })},
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
						d[7] = 3

						return d
					}),
				},
				{
					name: "non-monotonic fanout",
					data: corrupt(func(d []byte) []byte {
						binary.BigEndian.PutUint32(d[8:], 0xffffffff)

						return d
					}),
				},
				{
					name: "tables exceed index size",
					data: corrupt(func(d []byte) []byte {
						binary.BigEndian.PutUint32(d[8+255*4:], 0x00ffffff)

						return d
					}),
				},
				{
					name: "trailing size not 64-bit multiple",
					data: corrupt(func(d []byte) []byte { return append(d, 0xde, 0xad, 0xbe, 0xef) }),
				},
				{
					name: "more 64-bit offsets than objects",
					data: corrupt(func(d []byte) []byte {
						return append(d, bytes.Repeat([]byte{0x00}, 9*8)...)
					}),
				},
			}

			for _, tc := range cases {
				t.Run(tc.name, func(t *testing.T) {
					t.Parallel()

					_, err := packidx.Parse(tc.data, hashSize)
					if !errors.Is(err, packidx.ErrMalformedPackIndex) {
						t.Fatalf("Parse error = %v, want ErrMalformedPackIndex", err)
					}
				})
			}
		})
	}
}
