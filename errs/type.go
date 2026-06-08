package errs

import (
	"fmt"

	"lindenii.org/go/furgit/object/id"
	"lindenii.org/go/furgit/object/typ"
)

// ObjectTypeError indicates that a referenced object
// has a different type than what the operation expected.
type ObjectTypeError struct {
	OID  id.ObjectID
	Got  typ.Type
	Want typ.Type
}

// Error implements error.
func (e *ObjectTypeError) Error() string {
	return fmt.Sprintf(
		"object %s has type %s, want %s",
		e.OID, e.Got.Name(), e.Want.Name(),
	)
}
