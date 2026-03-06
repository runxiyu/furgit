package commitgraph

import (
	"fmt"

	"codeberg.org/lindenii/furgit/objectid"
)

// ErrNotFound reports a missing commit graph entry by object ID.
type ErrNotFound struct {
	OID objectid.ObjectID
}

// Error implements error.
func (err *ErrNotFound) Error() string {
	return fmt.Sprintf("format/commitgraph: object not found: %s", err.OID)
}

// ErrPositionOutOfRange reports an invalid graph position.
type ErrPositionOutOfRange struct {
	Pos Position
}

// Error implements error.
func (err *ErrPositionOutOfRange) Error() string {
	return fmt.Sprintf("format/commitgraph: position out of range: graph=%d index=%d", err.Pos.Graph, err.Pos.Index)
}

// ErrMalformed reports malformed commit-graph data.
type ErrMalformed struct {
	Path   string
	Reason string
}

// Error implements error.
func (err *ErrMalformed) Error() string {
	return fmt.Sprintf("format/commitgraph: malformed %q: %s", err.Path, err.Reason)
}

// ErrUnsupportedVersion reports unsupported commit-graph version.
type ErrUnsupportedVersion struct {
	Version uint8
}

// Error implements error.
func (err *ErrUnsupportedVersion) Error() string {
	return fmt.Sprintf("format/commitgraph: unsupported version %d", err.Version)
}

// ErrBloomUnavailable reports missing changed-path bloom data at one position.
type ErrBloomUnavailable struct {
	Pos Position
}

// Error implements error.
func (err *ErrBloomUnavailable) Error() string {
	return fmt.Sprintf("format/commitgraph: bloom unavailable at position graph=%d index=%d", err.Pos.Graph, err.Pos.Index)
}
