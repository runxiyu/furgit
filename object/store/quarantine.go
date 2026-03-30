package objectstore

// Quarantine is one quarantined write. It is intended to be embedded.
type Quarantine interface {
	// Reader returns the objects written into this quarantine.
	Reader

	// Promote publishes quarantined writes into their final destination.
	Promote() error

	// Discard abandons quarantined writes.
	Discard() error
}
