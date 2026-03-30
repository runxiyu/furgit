package objectstore

// BaseQuarantine is one quarantined write. It is intended to be embedded.
type BaseQuarantine interface {
	// Reader exposes the objects written into this quarantine.
	Reader

	// Promote publishes quarantined writes into their final destination.
	//
	// Promote invalidates the receiver.
	Promote() error

	// Discard abandons quarantined writes.
	//
	// Discard invalidates the receiver.
	Discard() error
}
