package reachability

import (
	"codeberg.org/lindenii/furgit/errors"
	objectcommit "codeberg.org/lindenii/furgit/object/commit"
	objectid "codeberg.org/lindenii/furgit/object/id"
	objecttype "codeberg.org/lindenii/furgit/object/type"
)

func (walk *Walk) validateCommitObject(id objectid.ObjectID) error {
	ty, err := walk.readHeaderType(id)
	if err != nil {
		return err
	}

	if ty != objecttype.TypeCommit {
		return &errors.ObjectTypeError{OID: id, Got: ty, Want: objecttype.TypeCommit}
	}

	content, err := walk.readBytesContent(id)
	if err != nil {
		return err
	}

	_, err = objectcommit.Parse(content, id.Algorithm())

	return err
}
