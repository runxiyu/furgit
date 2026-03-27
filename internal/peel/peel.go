// Package peel peels Git object references through annotated tags.
package peel

import (
	stderrors "errors"

	giterrors "codeberg.org/lindenii/furgit/errors"
	objectid "codeberg.org/lindenii/furgit/object/id"
	objectstore "codeberg.org/lindenii/furgit/object/store"
	"codeberg.org/lindenii/furgit/object/tag"
	objecttype "codeberg.org/lindenii/furgit/object/type"
)

// ToCommit peels annotated tags transitively until a commit is reached.
func ToCommit(store objectstore.Store, id objectid.ObjectID) (objectid.ObjectID, error) {
	for {
		ty, _, err := store.ReadHeader(id)
		if err != nil {
			if stderrors.Is(err, objectstore.ErrObjectNotFound) {
				return objectid.ObjectID{}, &giterrors.ObjectMissingError{OID: id}
			}

			return objectid.ObjectID{}, err
		}

		if ty != objecttype.TypeTag {
			if ty != objecttype.TypeCommit {
				return objectid.ObjectID{}, &giterrors.ObjectTypeError{OID: id, Got: ty, Want: objecttype.TypeCommit}
			}

			return id, nil
		}

		_, content, err := store.ReadBytesContent(id)
		if err != nil {
			if stderrors.Is(err, objectstore.ErrObjectNotFound) {
				return objectid.ObjectID{}, &giterrors.ObjectMissingError{OID: id}
			}

			return objectid.ObjectID{}, err
		}

		tag, err := tag.Parse(content, id.Algorithm())
		if err != nil {
			return objectid.ObjectID{}, err
		}

		id = tag.Target
	}
}
