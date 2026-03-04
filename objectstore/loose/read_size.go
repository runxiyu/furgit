package loose

import "codeberg.org/lindenii/furgit/objectid"

// ReadSize reads an object's declared content length.
func (store *Store) ReadSize(id objectid.ObjectID) (int64, error) {
	_, size, err := store.ReadHeader(id)

	return size, err
}
