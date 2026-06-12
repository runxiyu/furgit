package packrev

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"math"

	"lindenii.org/go/furgit/internal/stickyio"
	"lindenii.org/go/furgit/object/id"
)

// ErrInvalidPositions reports that
// positions supplied for a reverse index write
// are out of range or too numerous.
var ErrInvalidPositions = errors.New("internal/format/packrev: invalid positions")

// Write writes one pack reverse index to w.
//
// positions holds, for each object in pack offset order,
// the object's pack index position.
// packHash must be the pack's trailer hash;
// Write panics when its length does not match the object format.
func Write(w io.Writer, objectFormat id.ObjectFormat, positions []uint32, packHash []byte) error {
	hashID, err := hashFunctionID(objectFormat)
	if err != nil {
		return err
	}

	if len(packHash) != objectFormat.Size() {
		panic("internal/format/packrev: invalid pack hash length")
	}

	if len(positions) > math.MaxUint32 {
		return fmt.Errorf("%w: too many positions", ErrInvalidPositions)
	}

	for _, position := range positions {
		if uint64(position) >= uint64(len(positions)) {
			return fmt.Errorf("%w: index position out of range", ErrInvalidPositions)
		}
	}

	hashImpl, err := objectFormat.New()
	if err != nil {
		return fmt.Errorf("internal/format/packrev: %w", err)
	}

	bw := bufio.NewWriter(io.MultiWriter(w, hashImpl))
	sw := stickyio.New(bw)

	sw.PutUint32(signature)
	sw.PutUint32(version)
	sw.PutUint32(hashID)

	for _, position := range positions {
		sw.PutUint32(position)
	}

	sw.Put(packHash)

	err = sw.Err()
	if err != nil {
		return fmt.Errorf("internal/format/packrev: %w", err)
	}

	err = bw.Flush()
	if err != nil {
		return fmt.Errorf("internal/format/packrev: %w", err)
	}

	_, err = w.Write(hashImpl.Sum(nil))
	if err != nil {
		return fmt.Errorf("internal/format/packrev: %w", err)
	}

	return nil
}
