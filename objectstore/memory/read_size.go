package memory

import "codeberg.org/lindenii/furgit/objectid"

// ReadSize reads one object size.
func (store *Store) ReadSize(id objectid.ObjectID) (int64, error) {
	_, size, err := store.ReadHeader(id)
	if err != nil {
		return 0, err
	}

	return size, nil
}
