package fetch

import (
	"errors"

	"lindenii.org/go/furgit/errs"
	"lindenii.org/go/furgit/object/id"
	"lindenii.org/go/furgit/object/store"
)

// wrapObjectReadError maps raw object-store lookup failures to fetcher-level
// object lookup errors.
func wrapObjectReadError(id id.ObjectID, err error) error {
	if errors.Is(err, store.ErrObjectNotFound) {
		return &errs.ObjectMissingError{OID: id}
	}

	return err
}
