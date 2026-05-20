package refstore

import "errors"

// ErrReferenceNotFound indicates that a reference does not exist in a backend.
// TODO: Interface error? Just like object not found in objectstore.
// I'm not sure if I want this as a sentinel.
var ErrReferenceNotFound = errors.New("refstore: reference not found")
