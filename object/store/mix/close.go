package mix

// Close releases wrapper-local resources.
//
// Mix borrows its backends, so Close does not close them.
//
// Repeated calls to Close are undefined behavior.
func (mix *Mix) Close() error { return nil }
