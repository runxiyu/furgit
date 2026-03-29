package service

// Service executes protocol-independent receive-pack requests.
type Service struct {
	opts Options
}

// New creates one receive-pack service.
//
// Labels: Deps-Borrowed, Life-Parent.
func New(opts Options) *Service {
	return &Service{opts: opts}
}
