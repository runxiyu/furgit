package objectstore

// Quarantine is one quarantined write. It is intended to be embedded.
type Quarantine interface {
	// Reader returns the objects written into this quarantine.
	Reader() Reader

	// Promote publishes quarantined writes into their final destination.
	Promote() error

	// Discard abandons quarantined writes.
	Discard() error
}

// ObjectQuarantine represents one quarantined object-wise write.
type ObjectQuarantine interface {
	Quarantine
	ObjectWriter
}

// PackQuarantine represents one quarantined pack-wise write.
type PackQuarantine interface {
	Quarantine
	PackWriter
}

// ObjectQuarantineOptions controls the options for one object quarantine creation.
type ObjectQuarantineOptions struct{}

// PackQuarantineOptions controls the options for one pack quarantine creation.
type PackQuarantineOptions struct{}

// ObjectQuarantiner creates quarantines for object-wise writes.
type ObjectQuarantiner interface {
	BeginObjectQuarantine(opts ObjectQuarantineOptions) (ObjectQuarantine, error)
}

// PackQuarantiner creates quarantines for pack-wise writes.
type PackQuarantiner interface {
	BeginPackQuarantine(opts PackQuarantineOptions) (PackQuarantine, error)
}
