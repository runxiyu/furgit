package fetch

import (
	stderrors "errors"

	giterrors "codeberg.org/lindenii/furgit/errors"
	objectid "codeberg.org/lindenii/furgit/object/id"
	objectstore "codeberg.org/lindenii/furgit/object/store"
)

// wrapObjectReadError maps raw object-store lookup failures to fetcher-level
// object lookup errors.
func wrapObjectReadError(id objectid.ObjectID, err error) error {
	if stderrors.Is(err, objectstore.ErrObjectNotFound) {
		return &giterrors.ObjectMissingError{OID: id}
	}

	return err
}
