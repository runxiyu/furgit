package store

// QuarantineBase is one quarantined write.
// It is intended to be embedded.
type QuarantineBase interface {
	// Reader exposes the objects written into this quarantine.
	ObjectReader

	// Promote publishes quarantined writes into their final destination.
	//
	// Promote is not atomic.
	// Because objects are content-addressed and immutable,
	// a partial promotion is safe:
	// it only makes some objects visible early,
	// which is harmless until a ref references them.
	//
	// Promote invalidates the receiver.
	Promote() error

	// Discard abandons quarantined writes.
	//
	// Discard invalidates the receiver.
	Discard() error
}

// ObjectQuarantine represents one quarantined object-wise write.
type ObjectQuarantine interface {
	QuarantineBase
	ObjectWriter
}

// ObjectQuarantineOptions controls the options
// for one object quarantine creation.
type ObjectQuarantineOptions struct{}

// ObjectQuarantiner creates quarantines
// for object-wise writes.
type ObjectQuarantiner interface {
	BeginObjectQuarantine(opts ObjectQuarantineOptions) (ObjectQuarantine, error)
}

// PackQuarantine represents one quarantined pack-wise write.
type PackQuarantine interface {
	QuarantineBase
	PackWriter
}

// PackQuarantineOptions controls the options
// for one pack quarantine creation.
type PackQuarantineOptions struct{}

// PackQuarantiner creates quarantines
// for pack-wise writes.
type PackQuarantiner interface {
	BeginPackQuarantine(opts PackQuarantineOptions) (PackQuarantine, error)
}

// CoordinatedQuarantine represents one quarantine
// that accepts both object-wise and pack-wise writes
// behind a single Promote/Discard.
type CoordinatedQuarantine interface {
	ObjectQuarantine
	PackQuarantine
}

// CoordinatedQuarantineOptions controls the options
// for one coordinated quarantine creation.
type CoordinatedQuarantineOptions struct {
	Object ObjectQuarantineOptions
	Pack   PackQuarantineOptions
}

// CoordinatedQuarantiner creates coordinated quarantines
// that accept both object-wise and pack-wise writes
// behind a single Promote/Discard.
//
// The returned quarantine is usable
// anywhere either split quarantine is expected.
type CoordinatedQuarantiner interface {
	BeginCoordinatedQuarantine(opts CoordinatedQuarantineOptions) (CoordinatedQuarantine, error)
}
