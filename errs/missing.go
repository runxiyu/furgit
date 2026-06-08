package errs

import (
	"fmt"

	"lindenii.org/go/furgit/object/id"
)

// ObjectMissingError indicates that
// a referenced object is absent from the repository object store.
//
// This should only be used
// in situations where objects are being queried recursively
// or otherwise by some chain that the caller may not be aware of.
//
// Failures on direct object access
// should instead use [lindenii.org/go/furgit/object/store.ErrObjectNotFound].
type ObjectMissingError struct {
	OID id.ObjectID
}

// Error implements error.
func (e *ObjectMissingError) Error() string {
	return fmt.Sprintf("missing object %s", e.OID)
}
