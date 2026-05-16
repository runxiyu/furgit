package store

import "errors"

// ErrObjectNotFound indicates that
// an object does not exist in a backend.
// This error must only be produced by object stores,
// when it has no specified object ID,
// but no other unexpected conditions were encountered.
var ErrObjectNotFound = errors.New("objectstore: object not found")

// This is a sentinel with no details,
// because it could be a frequent occurence,
// and allocating frequently on expected error paths
// would be extremely harmful to performance.
// Sometime, I will audit this again.
// TODO