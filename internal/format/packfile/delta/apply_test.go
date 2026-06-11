package delta_test

import (
	"bytes"
	"errors"
	"testing"

	"lindenii.org/go/furgit/internal/format/packfile/delta"
)

func TestApplyInsert(t *testing.T) {
	t.Parallel()

	data := delta.AppendHeaderSizes(nil, 0, 5)
	data = append(data, 5, 'h', 'e', 'l', 'l', 'o')

	out, err := delta.Apply(nil, data)
	if err != nil {
		t.Fatalf("Apply: %v", err)
	}

	if string(out) != "hello" {
		t.Fatalf("Apply = %q, want %q", out, "hello")
	}
}

func TestApplyCopy(t *testing.T) {
	t.Parallel()

	base := []byte("0123456789")

	// One offset byte and one size byte.
	data := delta.AppendHeaderSizes(nil, 10, 4)
	data = append(data, 0x91, 3, 4)

	out, err := delta.Apply(base, data)
	if err != nil {
		t.Fatalf("Apply: %v", err)
	}

	if string(out) != "3456" {
		t.Fatalf("Apply = %q, want %q", out, "3456")
	}
}

func TestApplyCopyImplicitSize(t *testing.T) {
	t.Parallel()

	base := bytes.Repeat([]byte{0xab}, 0x10000+10)

	// Offset 0 and the implicit copy size 0x10000.
	data := delta.AppendHeaderSizes(nil, uint64(len(base)), 0x10000)
	data = append(data, 0x80)

	out, err := delta.Apply(base, data)
	if err != nil {
		t.Fatalf("Apply: %v", err)
	}

	if !bytes.Equal(out, base[:0x10000]) {
		t.Fatalf("Apply implicit-size copy mismatch")
	}
}

func TestApplyMixed(t *testing.T) {
	t.Parallel()

	base := []byte("the quick brown fox")

	data := delta.AppendHeaderSizes(nil, uint64(len(base)), 9)
	data = append(data, 0x91, 4, 5)       // copy "quick"
	data = append(data, 3, '!', '?', '!') // insert "!?!"
	data = append(data, 0x91, 16, 1)      // copy "f"

	out, err := delta.Apply(base, data)
	if err != nil {
		t.Fatalf("Apply: %v", err)
	}

	if string(out) != "quick!?!f" {
		t.Fatalf("Apply = %q, want %q", out, "quick!?!f")
	}
}

func TestApplyEmptyResult(t *testing.T) {
	t.Parallel()

	data := delta.AppendHeaderSizes(nil, 3, 0)

	out, err := delta.Apply([]byte("abc"), data)
	if err != nil {
		t.Fatalf("Apply: %v", err)
	}

	if len(out) != 0 {
		t.Fatalf("Apply = %q, want empty", out)
	}
}

func TestApplyMalformed(t *testing.T) {
	t.Parallel()

	base := []byte("0123456789")

	cases := []struct {
		name       string
		ops        []byte
		baseSize   uint64
		resultSize uint64
	}{
		{
			name:       "opcode zero",
			ops:        []byte{0x00},
			baseSize:   10,
			resultSize: 1,
		},
		{
			name:       "truncated copy operand",
			ops:        []byte{0x91, 3},
			baseSize:   10,
			resultSize: 4,
		},
		{
			name:       "copy past base",
			ops:        []byte{0x91, 8, 5},
			baseSize:   10,
			resultSize: 5,
		},
		{
			name:       "copy past result",
			ops:        []byte{0x91, 0, 5},
			baseSize:   10,
			resultSize: 3,
		},
		{
			name:       "insert past delta",
			ops:        []byte{5, 'a', 'b'},
			baseSize:   10,
			resultSize: 5,
		},
		{
			name:       "insert past result",
			ops:        []byte{5, 'a', 'b', 'c', 'd', 'e'},
			baseSize:   10,
			resultSize: 3,
		},
		{
			name:       "base size mismatch",
			ops:        []byte{0x91, 0, 1},
			baseSize:   9,
			resultSize: 1,
		},
		{
			name:       "result shorter than declared",
			ops:        []byte{0x91, 0, 1},
			baseSize:   10,
			resultSize: 2,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			data := delta.AppendHeaderSizes(nil, tc.baseSize, tc.resultSize)
			data = append(data, tc.ops...)

			_, err := delta.Apply(base, data)
			if !errors.Is(err, delta.ErrMalformedDelta) {
				t.Fatalf("Apply error = %v, want ErrMalformedDelta", err)
			}
		})
	}
}
