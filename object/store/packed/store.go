package packed

import (
	"codeberg.org/lindenii/furgit/object/store/packed/internal/reading"
)

// Store reads Git objects from pack/index files under an objects/pack root.
//
// Labels: Close-Caller.
type Store struct {
	reader *reading.Store
}

// Close releases mapped pack/index resources associated with the store.
func (store *Store) Close() error {
	return store.reader.Close()
}
