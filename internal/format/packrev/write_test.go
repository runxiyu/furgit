package packrev_test

import (
	"bytes"
	"errors"
	"os"
	"testing"

	"lindenii.org/go/furgit/internal/format/packidx"
	"lindenii.org/go/furgit/internal/format/packrev"
	"lindenii.org/go/furgit/object/id"
)

// writeSyntheticRev writes one reverse index over positions
// with a fixed fake pack hash.
func writeSyntheticRev(t *testing.T, objectFormat id.ObjectFormat, positions []uint32) []byte {
	t.Helper()

	packHash := bytes.Repeat([]byte{0x5a}, objectFormat.Size())

	var buf bytes.Buffer

	err := packrev.Write(&buf, objectFormat, positions, packHash)
	if err != nil {
		t.Fatalf("Write: %v", err)
	}

	return buf.Bytes()
}

func TestWriteRoundTrip(t *testing.T) {
	t.Parallel()

	for _, objectFormat := range id.SupportedObjectFormats() {
		t.Run(objectFormat.String(), func(t *testing.T) {
			t.Parallel()

			positions := []uint32{8, 6, 7, 5, 3, 0, 4, 1, 2}
			data := writeSyntheticRev(t, objectFormat, positions)

			rev, err := packrev.Parse(data, objectFormat)
			if err != nil {
				t.Fatalf("Parse: %v", err)
			}

			if rev.NumObjects() != len(positions) {
				t.Fatalf("NumObjects = %d, want %d", rev.NumObjects(), len(positions))
			}

			if !bytes.Equal(rev.PackHash(), bytes.Repeat([]byte{0x5a}, objectFormat.Size())) {
				t.Fatalf("PackHash mismatch")
			}

			for packOrder, want := range positions {
				position, err := rev.PositionAt(packOrder)
				if err != nil {
					t.Fatalf("PositionAt(%d): %v", packOrder, err)
				}

				if position != int(want) {
					t.Fatalf("PositionAt(%d) = %d, want %d", packOrder, position, want)
				}
			}
		})
	}
}

func TestWriteMatchesGit(t *testing.T) {
	t.Parallel()

	for _, objectFormat := range id.SupportedObjectFormats() {
		t.Run(objectFormat.String(), func(t *testing.T) {
			t.Parallel()

			prefix := makeGitPack(t, objectFormat)

			gitData, err := os.ReadFile(prefix + ".rev") //nolint:gosec
			if err != nil {
				t.Fatalf("ReadFile: %v", err)
			}

			idxData, err := os.ReadFile(prefix + ".idx") //nolint:gosec
			if err != nil {
				t.Fatalf("ReadFile: %v", err)
			}

			idx, err := packidx.Parse(idxData, objectFormat.Size())
			if err != nil {
				t.Fatalf("packidx.Parse: %v", err)
			}

			positions := packOrderPositions(t, &idx)

			var buf bytes.Buffer

			err = packrev.Write(&buf, objectFormat, positions, idx.PackHash())
			if err != nil {
				t.Fatalf("Write: %v", err)
			}

			if !bytes.Equal(buf.Bytes(), gitData) {
				t.Fatalf("Write output differs from git's reverse index (%d vs %d bytes)", buf.Len(), len(gitData))
			}
		})
	}
}

func TestWriteInvalidPositions(t *testing.T) {
	t.Parallel()

	for _, objectFormat := range id.SupportedObjectFormats() {
		t.Run(objectFormat.String(), func(t *testing.T) {
			t.Parallel()

			packHash := bytes.Repeat([]byte{0x5a}, objectFormat.Size())

			err := packrev.Write(&bytes.Buffer{}, objectFormat, []uint32{0, 5}, packHash)
			if !errors.Is(err, packrev.ErrInvalidPositions) {
				t.Fatalf("Write error = %v, want ErrInvalidPositions", err)
			}
		})
	}
}

func TestWriteBadPackHashPanics(t *testing.T) {
	t.Parallel()

	defer func() {
		if recover() == nil {
			t.Fatalf("Write with short pack hash: expected panic")
		}
	}()

	_ = packrev.Write(&bytes.Buffer{}, id.ObjectFormatSHA256, nil, []byte{0x01})
}
