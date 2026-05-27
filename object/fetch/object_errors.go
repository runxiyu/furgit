package fetch

import (
	stderrors "errors"

	giterrors "lindenii.org/go/furgit/errors"
	objectid "lindenii.org/go/furgit/object/id"
	objectstore "lindenii.org/go/furgit/object/store"
)

// wrapObjectReadError maps raw object-store lookup failures to fetcher-level
// object lookup errors.
func wrapObjectReadError(id objectid.ObjectID, err error) error {
	if stderrors.Is(err, objectstore.ErrObjectNotFound) {
		return &giterrors.ObjectMissingError{OID: id}
	}

	return err
}
