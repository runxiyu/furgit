package bloom_test

import (
	"encoding/binary"
	"errors"
	"testing"

	"lindenii.org/go/furgit/internal/format/packidx/bloom"
	"lindenii.org/go/furgit/object/id"
)

func validFilter(t *testing.T, format id.ObjectFormat) []byte {
	// TODO: maybe testgit should have something like this?
	t.Helper()

	builder, err := bloom.NewBuilder(format, 4, 2, make([]byte, format.Size()))
	if err != nil {
		t.Fatal(err)
	}

	return builder.Bytes()
}

func otherFormat(t *testing.T, format id.ObjectFormat) id.ObjectFormat {
	t.Helper()

	for _, candidate := range id.SupportedObjectFormats() {
		if candidate != format {
			return candidate
		}
	}

	t.Skip("only one supported object format")

	return id.ObjectFormatUnknown
}

func TestParseValid(t *testing.T) {
	t.Parallel()

	for _, format := range id.SupportedObjectFormats() {
		t.Run(format.String(), func(t *testing.T) {
			t.Parallel()

			_, err := bloom.Parse(validFilter(t, format), format)
			if err != nil {
				t.Fatalf("Parse rejected a valid filter: %v", err)
			}
		})
	}
}

func TestParseMalformed(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name   string
		mangle func(data []byte) []byte
	}{
		{"truncated", func(data []byte) []byte { return data[:bloom.HeaderLen-1] }},
		{"bad signature", func(data []byte) []byte {
			data[0] ^= 0xff

			return data
		}},
		{"bad version", func(data []byte) []byte {
			binary.BigEndian.PutUint32(data[4:], 99)

			return data
		}},
		{"non power of two", func(data []byte) []byte {
			binary.BigEndian.PutUint32(data[12:], 3)

			return data
		}},
		{"zero probe count", func(data []byte) []byte {
			binary.BigEndian.PutUint16(data[16:], 0)

			return data
		}},
		{"parameters exceed hash", func(data []byte) []byte {
			binary.BigEndian.PutUint32(data[12:], 1<<31)
			binary.BigEndian.PutUint16(data[16:], 30)

			return data
		}},
		{"nonzero padding", func(data []byte) []byte {
			data[20] = 1

			return data
		}},
		{"size disagrees", func(data []byte) []byte { return data[:len(data)-1] }},
	}

	for _, format := range id.SupportedObjectFormats() {
		t.Run(format.String(), func(t *testing.T) {
			t.Parallel()

			for _, tc := range cases {
				t.Run(tc.name, func(t *testing.T) {
					t.Parallel()

					data := tc.mangle(append([]byte(nil), validFilter(t, format)...))

					_, err := bloom.Parse(data, format)
					if !errors.Is(err, bloom.ErrMalformedBloomFilter) {
						t.Fatalf("Parse error = %v, want ErrMalformedBloomFilter", err)
					}
				})
			}
		})
	}
}

// TestVerifyDetectsCorruption checks that Verify accepts a sound filter
// and rejects one whose bucket bytes have been altered.
func TestVerifyDetectsCorruption(t *testing.T) {
	t.Parallel()

	for _, format := range id.SupportedObjectFormats() {
		t.Run(format.String(), func(t *testing.T) {
			t.Parallel()

			data := validFilter(t, format)

			filter, err := bloom.Parse(data, format)
			if err != nil {
				t.Fatal(err)
			}

			err = filter.Verify()
			if err != nil {
				t.Fatalf("Verify on a sound filter: %v", err)
			}

			data[bloom.HeaderLen] ^= 0xff

			err = filter.Verify()
			if !errors.Is(err, bloom.ErrMalformedBloomFilter) {
				t.Fatalf("Verify error = %v, want ErrMalformedBloomFilter", err)
			}
		})
	}
}

func TestParseHashMismatch(t *testing.T) {
	t.Parallel()

	for _, format := range id.SupportedObjectFormats() {
		t.Run(format.String(), func(t *testing.T) {
			t.Parallel()

			_, err := bloom.Parse(validFilter(t, format), otherFormat(t, format))
			if !errors.Is(err, bloom.ErrMalformedBloomFilter) {
				t.Fatalf("Parse error = %v, want ErrMalformedBloomFilter", err)
			}
		})
	}
}
