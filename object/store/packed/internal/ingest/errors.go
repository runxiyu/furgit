package ingest

import (
	"errors"
	"fmt"

	"lindenii.org/go/furgit/object/id"
)

// ErrMalformedPack reports that
// the incoming pack stream is truncated,
// inconsistent, or otherwise unparseable.
var ErrMalformedPack = errors.New("object/store/packed/internal/ingest: malformed pack")

// ErrThinPackNotPermitted reports that
// the incoming pack is thin,
// referencing bases not contained within it,
// but no external base reader was supplied to complete it.
var ErrThinPackNotPermitted = errors.New("object/store/packed/internal/ingest: thin pack not permitted: no thin base supplied")

// ThinBasesMissingError reports that
// an incoming thin pack references base objects
// that the supplied thin base reader does not contain,
// so the pack cannot be completed.
type ThinBasesMissingError struct {
	// OIDs holds the missing base object IDs, sorted.
	OIDs []id.ObjectID
}

// Error implements error.
func (e *ThinBasesMissingError) Error() string {
	return fmt.Sprintf(
		"object/store/packed/internal/ingest: thin pack references %d missing base objects",
		len(e.OIDs),
	)
}
