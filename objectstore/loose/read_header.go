package loose

import (
	"bufio"
	"compress/zlib"

	"codeberg.org/lindenii/furgit/objectid"
	"codeberg.org/lindenii/furgit/objecttype"
)

// ReadHeader reads an object's type and declared content length.
func (store *Store) ReadHeader(id objectid.ObjectID) (objecttype.Type, int64, error) {
	file, err := store.openObject(id)
	if err != nil {
		return objecttype.TypeInvalid, 0, err
	}
	defer func() { _ = file.Close() }()

	zr, err := zlib.NewReader(file)
	if err != nil {
		return objecttype.TypeInvalid, 0, err
	}
	defer func() { _ = zr.Close() }()

	ty, size, err := readHeader(bufio.NewReader(zr))
	if err != nil {
		return objecttype.TypeInvalid, 0, err
	}
	return ty, size, nil
}
