package packfile_test

import (
	"bytes"
	"errors"
	"math"
	"testing"

	"lindenii.org/go/furgit/internal/format/packfile"
)

func TestParseOfsDeltaDistance(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name     string
		data     []byte
		dist     uint64
		consumed int
	}{
		{name: "zero", data: []byte{0x00}, dist: 0, consumed: 1},
		{name: "one byte max", data: []byte{0x7f}, dist: 127, consumed: 1},
		{name: "two byte min", data: []byte{0x80, 0x00}, dist: 128, consumed: 2},
		{name: "two byte max", data: []byte{0xff, 0x7f}, dist: 16511, consumed: 2},
		{name: "trailing bytes ignored", data: []byte{0x7f, 0xde, 0xad}, dist: 127, consumed: 1},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			dist, consumed, err := packfile.ParseOfsDeltaDistance(tc.data)
			if err != nil {
				t.Fatalf("ParseOfsDeltaDistance: %v", err)
			}

			if dist != tc.dist {
				t.Fatalf("ParseOfsDeltaDistance = %d, want %d", dist, tc.dist)
			}

			if consumed != tc.consumed {
				t.Fatalf("ParseOfsDeltaDistance consumed = %d, want %d", consumed, tc.consumed)
			}
		})
	}
}

func TestParseOfsDeltaDistanceMalformed(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name string
		data []byte
	}{
		{name: "empty", data: []byte{}},
		{name: "truncated", data: []byte{0x80}},
		{name: "overflow", data: bytes.Repeat([]byte{0xff}, 10)},
		{name: "overlong", data: append(bytes.Repeat([]byte{0x80}, 10), 0x00)},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			_, _, err := packfile.ParseOfsDeltaDistance(tc.data)
			if !errors.Is(err, packfile.ErrMalformedOfsDeltaDistance) {
				t.Fatalf("ParseOfsDeltaDistance error = %v, want ErrMalformedOfsDeltaDistance", err)
			}
		})
	}
}

func TestOfsDeltaDistanceRoundTrip(t *testing.T) {
	t.Parallel()

	distances := []uint64{
		0, 1, 127, 128, 16511, 16512, 1 << 31, 1 << 57, math.MaxUint64,
	}

	for _, dist := range distances {
		data := packfile.AppendOfsDeltaDistance(nil, dist)
		if len(data) > packfile.MaxOfsDeltaDistanceLen {
			t.Fatalf("AppendOfsDeltaDistance(%d) length = %d, want <= %d",
				dist, len(data), packfile.MaxOfsDeltaDistanceLen)
		}

		got, consumed, err := packfile.ParseOfsDeltaDistance(data)
		if err != nil {
			t.Fatalf("ParseOfsDeltaDistance: %v", err)
		}

		if got != dist {
			t.Fatalf("round trip = %d, want %d", got, dist)
		}

		if consumed != len(data) {
			t.Fatalf("round trip consumed = %d, want %d", consumed, len(data))
		}
	}
}
