package reachability

import (
	"codeberg.org/lindenii/furgit/object"
	"codeberg.org/lindenii/furgit/objectid"
	"codeberg.org/lindenii/furgit/objecttype"
)

// peelRootToCommit peels annotated tags transitively until a commit is reached.
func (r *Reachability) peelRootToCommit(id objectid.ObjectID) (objectid.ObjectID, error) {
	for {
		ty, err := r.readHeaderType(id)
		if err != nil {
			return objectid.ObjectID{}, err
		}

		if ty != objecttype.TypeTag {
			if ty != objecttype.TypeCommit {
				return objectid.ObjectID{}, &ObjectTypeError{OID: id, Got: ty, Want: objecttype.TypeCommit}
			}

			return id, nil
		}

		content, err := r.readBytesContent(id)
		if err != nil {
			return objectid.ObjectID{}, err
		}

		tag, err := object.ParseTag(content, id.Algorithm())
		if err != nil {
			return objectid.ObjectID{}, err
		}

		id = tag.Target
	}
}
