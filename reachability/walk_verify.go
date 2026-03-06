package reachability

import (
	"codeberg.org/lindenii/furgit/object"
	"codeberg.org/lindenii/furgit/objectid"
	"codeberg.org/lindenii/furgit/objecttype"
)

func (walk *Walk) validateCommitObject(id objectid.ObjectID) error {
	ty, err := walk.readHeaderType(id)
	if err != nil {
		return err
	}

	if ty != objecttype.TypeCommit {
		return &ErrObjectType{OID: id, Got: ty, Want: objecttype.TypeCommit}
	}

	content, err := walk.readBytesContent(id)
	if err != nil {
		return err
	}

	_, err = object.ParseCommit(content, id.Algorithm())

	return err
}
