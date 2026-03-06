package errors

import (
	"fmt"

	"codeberg.org/lindenii/furgit/objectid"
)

// ObjectMissingError indicates that a referenced object is absent from the
// repository object store.
type ObjectMissingError struct {
	OID objectid.ObjectID
}

// Error implements error.
func (e *ObjectMissingError) Error() string {
	return fmt.Sprintf("missing object %s", e.OID)
}
