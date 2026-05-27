package memory

import objectid "lindenii.org/go/furgit/object/id"

// ReadSize reads one object size.
func (store *Store) ReadSize(id objectid.ObjectID) (int64, error) {
	_, size, err := store.ReadHeader(id)
	if err != nil {
		return 0, err
	}

	return size, nil
}
