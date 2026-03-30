package objectstore

import "io"

// PackWriteOptions controls one pack write operation.
type PackWriteOptions struct{}

// PackWriter writes Git pack streams.
type PackWriter interface {
	// WritePack ingests one pack stream.
	WritePack(src io.Reader, opts PackWriteOptions) error
}

// PackQuarantine represents one quarantined pack-wise write.
type PackQuarantine interface {
	Quarantine
	PackWriter
}

// PackQuarantineOptions controls the options for one pack quarantine creation.
type PackQuarantineOptions struct{}

// PackQuarantiner creates quarantines for pack-wise writes.
type PackQuarantiner interface {
	BeginPackQuarantine(opts PackQuarantineOptions) (PackQuarantine, error)
}
