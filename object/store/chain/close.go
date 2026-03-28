package chain

// Close releases wrapper-local resources.
//
// Chain borrows its backends, so Close does not close them.
//
// Repeated calls to Close are undefined behavior.
func (chain *Chain) Close() error { return nil }
