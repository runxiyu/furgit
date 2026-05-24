package id

import "errors"

var (
	// ErrInvalidObjectFormat indicates an unsupported object format.
	ErrInvalidObjectFormat = errors.New("object/id: invalid object format")

	// ErrInvalidObjectID indicates malformed object ID data.
	ErrInvalidObjectID = errors.New("object/id: invalid object id")
)
