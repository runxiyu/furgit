package packed

import (
	"os"

	objectid "codeberg.org/lindenii/furgit/object/id"
	"codeberg.org/lindenii/furgit/object/store/packed/internal/reading"
)

// New creates a packed-object store rooted at an objects/pack directory.
//
// Labels: Deps-Borrowed, Life-Parent.
func New(root *os.Root, algo objectid.Algorithm, opts Options) (*Store, error) {
	reader, err := reading.New(root, algo, opts.toReadingOptions())
	if err != nil {
		return nil, err
	}

	return &Store{reader: reader}, nil
}
