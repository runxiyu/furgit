package service

// Service executes protocol-independent receive-pack requests.
//
// Service borrows all dependencies supplied in Options.
type Service struct {
	opts Options
}

// New creates one receive-pack service.
//
// The returned service borrows opts and does not take ownership of any stores,
// roots, hooks, or I/O endpoints reachable through it.
func New(opts Options) *Service {
	return &Service{opts: opts}
}
