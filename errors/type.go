package errors

import (
	"fmt"

	objectid "codeberg.org/lindenii/furgit/object/id"
	objecttype "codeberg.org/lindenii/furgit/object/type"
)

// ObjectTypeError indicates that a referenced object has a different type than
// what the operation expected.
type ObjectTypeError struct {
	OID  objectid.ObjectID
	Got  objecttype.Type
	Want objecttype.Type
}

// Error implements error.
func (e *ObjectTypeError) Error() string {
	gotName, gotOK := objecttype.Name(e.Got)
	if !gotOK {
		gotName = fmt.Sprintf("type(%d)", e.Got)
	}

	wantName, wantOK := objecttype.Name(e.Want)
	if !wantOK {
		wantName = fmt.Sprintf("type(%d)", e.Want)
	}

	return fmt.Sprintf("object %s has type %s, want %s", e.OID, gotName, wantName)
}
