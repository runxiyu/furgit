package packfile_test

import (
	"bytes"
	"errors"
	"math"
	"testing"

	"lindenii.org/go/furgit/internal/format/packfile"
	"lindenii.org/go/furgit/object/id"
)

func TestParseEntryHeaderBase(t *testing.T) {
	t.Parallel()

	// Commit entry of size 85:
	// 0x95 has the continuation bit set, type 1, and size low bits 0x5;
	// 0x05 contributes 5<<4.
	header, err := packfile.ParseEntryHeader([]byte{0x95, 0x05}, 20)
	if err != nil {
		t.Fatalf("ParseEntryHeader: %v", err)
	}

	if header.Type != packfile.EntryTypeCommit {
		t.Fatalf("ParseEntryHeader type = %d, want %d", header.Type, packfile.EntryTypeCommit)
	}

	if header.Size != 85 {
		t.Fatalf("ParseEntryHeader size = %d, want 85", header.Size)
	}

	if header.HeaderLen != 2 {
		t.Fatalf("ParseEntryHeader header len = %d, want 2", header.HeaderLen)
	}
}

func TestParseEntryHeaderRefDelta(t *testing.T) {
	t.Parallel()

	for _, objectFormat := range id.SupportedObjectFormats() {
		t.Run(objectFormat.String(), func(t *testing.T) {
			t.Parallel()

			hashSize := objectFormat.Size()

			base := bytes.Repeat([]byte{0xab}, hashSize)
			data := append([]byte{0x73}, base...)

			header, err := packfile.ParseEntryHeader(data, hashSize)
			if err != nil {
				t.Fatalf("ParseEntryHeader: %v", err)
			}

			if header.Type != packfile.EntryTypeRefDelta {
				t.Fatalf("ParseEntryHeader type = %d, want %d", header.Type, packfile.EntryTypeRefDelta)
			}

			if header.Size != 3 {
				t.Fatalf("ParseEntryHeader size = %d, want 3", header.Size)
			}

			if header.HeaderLen != 1+hashSize {
				t.Fatalf("ParseEntryHeader header len = %d, want %d", header.HeaderLen, 1+hashSize)
			}

			if !bytes.Equal(header.RefBase[:hashSize], base) {
				t.Fatalf("ParseEntryHeader ref base mismatch")
			}
		})
	}
}

func TestParseEntryHeaderOfsDelta(t *testing.T) {
	t.Parallel()

	header, err := packfile.ParseEntryHeader([]byte{0x65, 0x80, 0x00}, 20)
	if err != nil {
		t.Fatalf("ParseEntryHeader: %v", err)
	}

	if header.Type != packfile.EntryTypeOfsDelta {
		t.Fatalf("ParseEntryHeader type = %d, want %d", header.Type, packfile.EntryTypeOfsDelta)
	}

	if header.OfsDistance != 128 {
		t.Fatalf("ParseEntryHeader distance = %d, want 128", header.OfsDistance)
	}

	if header.HeaderLen != 3 {
		t.Fatalf("ParseEntryHeader header len = %d, want 3", header.HeaderLen)
	}
}

func TestParseEntryHeaderMalformed(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name string
		data []byte
	}{
		{name: "empty", data: []byte{}},
		{name: "truncated type/size", data: []byte{0x95}},
		{
			name: "size overflow",
			data: append([]byte{0x9f}, bytes.Repeat([]byte{0xff}, 8)...),
		},
		{
			name: "overlong type/size",
			data: append([]byte{0x95}, append(bytes.Repeat([]byte{0x80}, 9), 0x00)...),
		},
		{name: "truncated ref-delta base", data: []byte{0x73, 0xab, 0xab}},
		{name: "truncated ofs distance", data: []byte{0x65, 0x80}},
		{name: "type invalid", data: []byte{0x05}},
		{name: "type future", data: []byte{0x55}},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			_, err := packfile.ParseEntryHeader(tc.data, 20)
			if !errors.Is(err, packfile.ErrMalformedEntryHeader) {
				t.Fatalf("ParseEntryHeader error = %v, want ErrMalformedEntryHeader", err)
			}
		})
	}
}

func TestParseEntryHeaderBadHashSize(t *testing.T) {
	t.Parallel()

	for _, hashSize := range []int{-1, 0, id.MaxObjectIDSize + 1} {
		_, err := packfile.ParseEntryHeader([]byte{0x95, 0x05}, hashSize)
		if err == nil {
			t.Fatalf("ParseEntryHeader hash size %d: expected error", hashSize)
		}
	}
}

func TestTypeSizeRoundTrip(t *testing.T) {
	t.Parallel()

	baseTypes := []packfile.EntryType{
		packfile.EntryTypeCommit,
		packfile.EntryTypeTree,
		packfile.EntryTypeBlob,
		packfile.EntryTypeTag,
	}

	sizes := []uint64{
		0, 1, 15, 16, 127, 128, 1 << 20, 1 << 57, math.MaxUint64,
	}

	for _, entryType := range baseTypes {
		for _, size := range sizes {
			data := packfile.AppendTypeSize(nil, entryType, size)
			if len(data) > packfile.MaxTypeSizeLen {
				t.Fatalf("AppendTypeSize(%d, %d) length = %d, want <= %d",
					entryType, size, len(data), packfile.MaxTypeSizeLen)
			}

			header, err := packfile.ParseEntryHeader(data, 20)
			if err != nil {
				t.Fatalf("ParseEntryHeader: %v", err)
			}

			if header.Type != entryType {
				t.Fatalf("round trip type = %d, want %d", header.Type, entryType)
			}

			if header.Size != size {
				t.Fatalf("round trip size = %d, want %d", header.Size, size)
			}

			if header.HeaderLen != len(data) {
				t.Fatalf("round trip header len = %d, want %d", header.HeaderLen, len(data))
			}
		}
	}
}

func TestRefDeltaHeaderRoundTrip(t *testing.T) {
	t.Parallel()

	for _, objectFormat := range id.SupportedObjectFormats() {
		t.Run(objectFormat.String(), func(t *testing.T) {
			t.Parallel()

			hashSize := objectFormat.Size()

			base := bytes.Repeat([]byte{0xcd}, hashSize)
			data := packfile.AppendTypeSize(nil, packfile.EntryTypeRefDelta, 42)
			data = append(data, base...)

			header, err := packfile.ParseEntryHeader(data, hashSize)
			if err != nil {
				t.Fatalf("ParseEntryHeader: %v", err)
			}

			if !bytes.Equal(header.RefBase[:hashSize], base) {
				t.Fatalf("round trip ref base mismatch")
			}

			if header.HeaderLen != len(data) {
				t.Fatalf("round trip header len = %d, want %d", header.HeaderLen, len(data))
			}
		})
	}
}

func TestOfsDeltaHeaderRoundTrip(t *testing.T) {
	t.Parallel()

	data := packfile.AppendTypeSize(nil, packfile.EntryTypeOfsDelta, 7)
	data = packfile.AppendOfsDeltaDistance(data, 123456)

	header, err := packfile.ParseEntryHeader(data, 20)
	if err != nil {
		t.Fatalf("ParseEntryHeader: %v", err)
	}

	if header.OfsDistance != 123456 {
		t.Fatalf("round trip distance = %d, want 123456", header.OfsDistance)
	}

	if header.HeaderLen != len(data) {
		t.Fatalf("round trip header len = %d, want %d", header.HeaderLen, len(data))
	}
}

func TestMaxEntryHeaderLenBounds(t *testing.T) {
	t.Parallel()

	for _, objectFormat := range id.SupportedObjectFormats() {
		hashSize := objectFormat.Size()

		maxLen := packfile.MaxEntryHeaderLen(hashSize)
		if maxLen < packfile.MaxTypeSizeLen+hashSize {
			t.Fatalf("MaxEntryHeaderLen(%d) = %d cannot fit a ref-delta header", hashSize, maxLen)
		}

		if maxLen < packfile.MaxTypeSizeLen+packfile.MaxOfsDeltaDistanceLen {
			t.Fatalf("MaxEntryHeaderLen(%d) = %d cannot fit an ofs-delta header", hashSize, maxLen)
		}
	}
}
