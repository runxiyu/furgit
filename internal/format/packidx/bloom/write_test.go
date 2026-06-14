package bloom_test

import (
	"bytes"
	"errors"
	"testing"

	"lindenii.org/go/furgit/internal/format/packidx/bloom"
	"lindenii.org/go/furgit/object/id"
)

func TestRecommendParams(t *testing.T) {
	t.Parallel()

	for _, format := range id.SupportedObjectFormats() {
		t.Run(format.String(), func(t *testing.T) {
			t.Parallel()

			for _, n := range []int{0, 1, 1000, 10000, 1000000} {
				bucketCount, k, err := bloom.RecommendParams(format, n)
				if err != nil {
					t.Fatalf("n=%d: %v", n, err)
				}

				if bucketCount == 0 || bucketCount&(bucketCount-1) != 0 {
					t.Errorf("n=%d: bucket count %d not a power of two", n, bucketCount)
				}

				_, err = bloom.NewBuilder(format, bucketCount, k)
				if err != nil {
					t.Errorf("n=%d: recommended parameters rejected: %v", n, err)
				}
			}
		})
	}
}

func TestNewBuilderRejects(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name        string
		bucketCount uint32
		k           uint16
	}{
		{"zero buckets", 0, 8},
		{"non power of two", 3, 8},
		{"zero probe count", 4, 0},
	}

	for _, format := range id.SupportedObjectFormats() {
		t.Run(format.String(), func(t *testing.T) {
			t.Parallel()

			for _, tc := range cases {
				t.Run(tc.name, func(t *testing.T) {
					t.Parallel()

					_, err := bloom.NewBuilder(format, tc.bucketCount, tc.k)
					if !errors.Is(err, bloom.ErrInvalidParameters) {
						t.Fatalf("error = %v, want ErrInvalidParameters", err)
					}
				})
			}
		})
	}
}

func TestAddBadLength(t *testing.T) {
	t.Parallel()

	builder, err := bloom.NewBuilder(id.ObjectFormatSHA256, 4, 2)
	if err != nil {
		t.Fatal(err)
	}

	defer func() {
		if recover() == nil {
			t.Fatal("Add did not panic on a short object ID")
		}
	}()

	builder.Add(make([]byte, id.ObjectFormatSHA256.Size()-1))
}

func TestWriteTo(t *testing.T) {
	t.Parallel()

	builder, err := bloom.NewBuilder(id.ObjectFormatSHA256, 4, 2)
	if err != nil {
		t.Fatal(err)
	}

	var buf bytes.Buffer

	n, err := builder.WriteTo(&buf)
	if err != nil {
		t.Fatal(err)
	}

	if int(n) != len(builder.Bytes()) {
		t.Fatalf("WriteTo wrote %d bytes, want %d", n, len(builder.Bytes()))
	}

	if !bytes.Equal(buf.Bytes(), builder.Bytes()) {
		t.Fatal("WriteTo output differs from Bytes")
	}
}
