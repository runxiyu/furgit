package reachability

import (
	"codeberg.org/lindenii/furgit/errors"
	objectid "codeberg.org/lindenii/furgit/object/id"
	objecttype "codeberg.org/lindenii/furgit/object/type"
)

func (walk *Walk) validateCommitObject(id objectid.ObjectID) error {
	ty, _, err := walk.reachability.fetcher.Header(id)
	if err != nil {
		return err
	}

	if ty != objecttype.TypeCommit {
		return &errors.ObjectTypeError{OID: id, Got: ty, Want: objecttype.TypeCommit}
	}

	_, err = walk.reachability.fetcher.ExactCommit(id)

	return err
}
