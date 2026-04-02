package id

import "errors"

var (
	// ErrInvalidAlgorithm indicates an unsupported object ID algorithm.
	ErrInvalidAlgorithm = errors.New("objectid: invalid algorithm")

	// ErrInvalidObjectID indicates malformed object ID data.
	ErrInvalidObjectID = errors.New("objectid: invalid object id")
)
