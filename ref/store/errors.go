package refstore

import "errors"

// ErrReferenceNotFound indicates that a reference does not exist in a backend.
// TODO: Interface error? Just like object not found in objectstorer.
var ErrReferenceNotFound = errors.New("refstore: reference not found")
