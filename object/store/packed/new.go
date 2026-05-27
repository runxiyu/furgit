package packed

import (
	"os"

	objectid "lindenii.org/go/furgit/object/id"
	"lindenii.org/go/furgit/object/store/packed/internal/reading"
)

// New creates a packed-object store rooted at an objects/pack directory.
//
// Labels: Deps-Borrowed, Life-Parent.
func New(root *os.Root, algo objectid.Algorithm, opts Options) (*Store, error) {
	reader, err := reading.New(root, algo, opts.toReadingOptions())
	if err != nil {
		return nil, err
	}

	return &Store{
		root:   root,
		algo:   algo,
		opts:   opts,
		reader: reader,
	}, nil
}
