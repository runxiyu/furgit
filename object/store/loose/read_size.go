package loose

import objectid "codeberg.org/lindenii/furgit/object/id"

// ReadSize reads an object's declared content length.
//
// Like ReadHeader, it parses only enough of the zlib-decoded object to recover
// the header and does not verify the zlib Adler-32 trailer.
func (store *Store) ReadSize(id objectid.ObjectID) (int64, error) {
	_, size, err := store.ReadHeader(id)

	return size, err
}
