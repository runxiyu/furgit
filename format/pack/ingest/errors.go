package ingest

import "fmt"

// ErrInvalidPackHeader reports an invalid or unsupported pack header.
type ErrInvalidPackHeader struct {
	Reason string
}

// Error implements error.
func (err *ErrInvalidPackHeader) Error() string {
	return fmt.Sprintf("format/pack/ingest: invalid pack header: %s", err.Reason)
}

// ErrPackTrailerMismatch reports a mismatch between computed and trailer pack hash.
type ErrPackTrailerMismatch struct{}

// Error implements error.
func (err *ErrPackTrailerMismatch) Error() string {
	return "format/pack/ingest: pack trailer hash mismatch"
}

// ErrThinPackUnresolved reports unresolved REF deltas when fixThin is disabled
// or when required bases cannot be found in base.
type ErrThinPackUnresolved struct {
	Count int
}

// Error implements error.
func (err *ErrThinPackUnresolved) Error() string {
	return fmt.Sprintf("format/pack/ingest: unresolved thin deltas: %d", err.Count)
}

// ErrMalformedPackEntry reports malformed entry encoding at one pack offset.
type ErrMalformedPackEntry struct {
	Offset uint64
	Reason string
}

// Error implements error.
func (err *ErrMalformedPackEntry) Error() string {
	return fmt.Sprintf("format/pack/ingest: malformed pack entry at offset %d: %s", err.Offset, err.Reason)
}

// ErrDeltaCycle reports a detected cycle in delta dependency resolution.
type ErrDeltaCycle struct {
	Offset uint64
}

// Error implements error.
func (err *ErrDeltaCycle) Error() string {
	return fmt.Sprintf("format/pack/ingest: delta cycle detected at offset %d", err.Offset)
}

// ErrDestinationWrite reports destination I/O failures.
type ErrDestinationWrite struct {
	Op string
}

// Error implements error.
func (err *ErrDestinationWrite) Error() string {
	return fmt.Sprintf("format/pack/ingest: destination write failure: %s", err.Op)
}
