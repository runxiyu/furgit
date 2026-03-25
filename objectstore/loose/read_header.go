package loose

import (
	"bufio"

	"codeberg.org/lindenii/furgit/internal/compress/zlib"
	"codeberg.org/lindenii/furgit/objectid"
	objecttype "codeberg.org/lindenii/furgit/object/type"
)

// ReadHeader reads an object's type and declared content length.
//
// It parses only enough of the zlib-decoded object to recover the object
// header. It does not verify that the remaining object content is readable and
// does not verify the zlib Adler-32 trailer.
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

	_, ty, size, err := readHeader(bufio.NewReader(zr))
	if err != nil {
		return objecttype.TypeInvalid, 0, err
	}

	return ty, size, nil
}
