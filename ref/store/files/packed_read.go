package files

import (
	"errors"
	"fmt"
	"os"

	"lindenii.org/go/furgit/ref"
)

func (store *Store) readPackedRefs() (*packedRefs, error) {
	file, err := store.commonRoot.Open("packed-refs")
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return &packedRefs{
				byName:  make(map[string]ref.Detached),
				ordered: nil,
			}, nil
		}

		return nil, fmt.Errorf("refstore/files: open packed-refs: %w", err)
	}

	defer func() { _ = file.Close() }()

	byName, ordered, err := parsePackedRefs(file, store.algo)
	if err != nil {
		return nil, err
	}

	return &packedRefs{
		byName:  byName,
		ordered: ordered,
	}, nil
}
