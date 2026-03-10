package ingest

import (
	"errors"
	"fmt"
)

// InvalidPackHeaderError reports an invalid or unsupported pack header.
type InvalidPackHeaderError struct {
	Reason string
}

// Error implements error.
func (err *InvalidPackHeaderError) Error() string {
	return "packfile/ingest: invalid pack header: " + err.Reason
}

// PackTrailerMismatchError reports a mismatch between computed and trailer pack hash.
type PackTrailerMismatchError struct{}

// Error implements error.
func (err *PackTrailerMismatchError) Error() string {
	return "packfile/ingest: pack trailer hash mismatch"
}

// ThinPackUnresolvedError reports unresolved REF deltas when fixThin is disabled
// or when required bases cannot be found in base.
type ThinPackUnresolvedError struct {
	Count int
}

// Error implements error.
func (err *ThinPackUnresolvedError) Error() string {
	return fmt.Sprintf("packfile/ingest: unresolved thin deltas: %d", err.Count)
}

// MalformedPackEntryError reports malformed entry encoding at one pack offset.
type MalformedPackEntryError struct {
	Offset uint64
	Reason string
}

// Error implements error.
func (err *MalformedPackEntryError) Error() string {
	return fmt.Sprintf("packfile/ingest: malformed pack entry at offset %d: %s", err.Offset, err.Reason)
}

// DeltaCycleError reports a detected cycle in delta dependency resolution.
type DeltaCycleError struct {
	Offset uint64
}

// Error implements error.
func (err *DeltaCycleError) Error() string {
	return fmt.Sprintf("packfile/ingest: delta cycle detected at offset %d", err.Offset)
}

// DestinationWriteError reports destination I/O failures.
type DestinationWriteError struct {
	Op string
}

// Error implements error.
func (err *DestinationWriteError) Error() string {
	return "packfile/ingest: destination write failure: " + err.Op
}

var errExternalThinBase = errors.New("packfile/ingest: external thin base required")

var (
	// ErrAlreadyFinalized indicates Continue/Discard already called.
	ErrAlreadyFinalized = errors.New("packfile/ingest: operation already finalized")
	// ErrZeroObjectContinue indicates Continue was called for a zero-object pack.
	ErrZeroObjectContinue = errors.New("packfile/ingest: cannot continue zero-object pack")
	// ErrNonZeroDiscard indicates Discard was called for a non-zero-object pack.
	ErrNonZeroDiscard = errors.New("packfile/ingest: cannot discard non-zero pack")
)
