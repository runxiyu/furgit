package objectstore

import "errors"

// ErrObjectNotFound indicates that an object does not exist in a backend.
// This error MUST only be used in situations where the object store has
// no specified object ID, but no other unexpected conditions were
// encountered. In particular, it is not suitable for situations where one
// object references another (such as a tree referencing a blob) but
// the latter does not exist; these situations should use a separate
// error (TODO).
var ErrObjectNotFound = errors.New("objectstore: object not found")
