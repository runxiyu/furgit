package read

import (
	"fmt"

	objectid "lindenii.org/go/furgit/object/id"
)

// NotFoundError reports a missing commit graph entry by object ID.
type NotFoundError struct {
	OID objectid.ObjectID
}

// Error implements error.
func (err *NotFoundError) Error() string {
	return fmt.Sprintf("commitgraph: object not found: %s", err.OID)
}

// PositionOutOfRangeError reports an invalid graph position.
type PositionOutOfRangeError struct {
	Pos Position
}

// Error implements error.
func (err *PositionOutOfRangeError) Error() string {
	return fmt.Sprintf("commitgraph: position out of range: graph=%d index=%d", err.Pos.Graph, err.Pos.Index)
}

// MalformedError reports malformed commit-graph data.
type MalformedError struct {
	Path   string
	Reason string
}

// Error implements error.
func (err *MalformedError) Error() string {
	return fmt.Sprintf("commitgraph: malformed %q: %s", err.Path, err.Reason)
}

// UnsupportedVersionError reports unsupported commit-graph version.
type UnsupportedVersionError struct {
	Version uint8
}

// Error implements error.
func (err *UnsupportedVersionError) Error() string {
	return fmt.Sprintf("commitgraph: unsupported version %d", err.Version)
}

// BloomUnavailableError reports missing changed-path bloom data at one position.
type BloomUnavailableError struct {
	Pos Position
}

// Error implements error.
func (err *BloomUnavailableError) Error() string {
	return fmt.Sprintf("commitgraph: bloom unavailable at position graph=%d index=%d", err.Pos.Graph, err.Pos.Index)
}
