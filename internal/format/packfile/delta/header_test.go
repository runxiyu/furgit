package delta_test

import (
	"bytes"
	"errors"
	"math"
	"testing"

	"lindenii.org/go/furgit/internal/format/packfile/delta"
)

func TestParseHeaderSizesRoundTrip(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		baseSize   uint64
		resultSize uint64
	}{
		{name: "zero", baseSize: 0, resultSize: 0},
		{name: "small", baseSize: 5, resultSize: 130},
		{name: "boundaries", baseSize: 127, resultSize: 128},
		{name: "large", baseSize: 1 << 32, resultSize: 1 << 57},
		{name: "max", baseSize: math.MaxUint64, resultSize: math.MaxUint64},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			data := delta.AppendHeaderSizes(nil, tc.baseSize, tc.resultSize)

			wantConsumed := len(data)

			data = append(data, 0xde, 0xad)

			baseSize, resultSize, consumed, err := delta.ParseHeaderSizes(data)
			if err != nil {
				t.Fatalf("ParseHeaderSizes: %v", err)
			}

			if baseSize != tc.baseSize {
				t.Fatalf("ParseHeaderSizes base size = %d, want %d", baseSize, tc.baseSize)
			}

			if resultSize != tc.resultSize {
				t.Fatalf("ParseHeaderSizes result size = %d, want %d", resultSize, tc.resultSize)
			}

			if consumed != wantConsumed {
				t.Fatalf("ParseHeaderSizes consumed = %d, want %d", consumed, wantConsumed)
			}
		})
	}
}

func TestParseHeaderSizesMalformed(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name string
		data []byte
	}{
		{name: "empty", data: []byte{}},
		{name: "truncated first varint", data: []byte{0x80}},
		{name: "missing second varint", data: []byte{0x05}},
		{name: "truncated second varint", data: []byte{0x05, 0x80}},
		{
			name: "overflow",
			data: append(bytes.Repeat([]byte{0xff}, 9), 0x7f),
		},
		{
			name: "overlong",
			data: append(bytes.Repeat([]byte{0x80}, 10), 0x00),
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			_, _, _, err := delta.ParseHeaderSizes(tc.data)
			if !errors.Is(err, delta.ErrMalformedDelta) {
				t.Fatalf("ParseHeaderSizes error = %v, want ErrMalformedDelta", err)
			}
		})
	}
}
