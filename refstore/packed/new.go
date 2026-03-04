package packed

import (
	"fmt"
	"os"

	"codeberg.org/lindenii/furgit/objectid"
)

// New parses packed-refs from one repository root using the given object ID
// algorithm.
func New(root *os.Root, algo objectid.Algorithm) (*Store, error) {
	if algo.Size() == 0 {
		return nil, objectid.ErrInvalidAlgorithm
	}

	packedRefs, err := root.Open("packed-refs")
	if err != nil {
		return nil, fmt.Errorf("refstore/packed: open packed-refs: %w", err)
	}

	defer func() { _ = packedRefs.Close() }()

	byName, ordered, err := parsePackedRefs(packedRefs, algo)
	if err != nil {
		return nil, err
	}

	return &Store{
		byName:  byName,
		ordered: ordered,
	}, nil
}
