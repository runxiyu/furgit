package receivepack

import "errors"

var (
	// ErrMissingAlgorithm reports one missing repository hash algorithm.
	ErrMissingAlgorithm = errors.New("receivepack: missing object id algorithm")
	// ErrMissingRefs reports one missing reference store dependency.
	ErrMissingRefs = errors.New("receivepack: missing refs store")
	// ErrMissingObjects reports one missing object store dependency.
	ErrMissingObjects = errors.New("receivepack: missing objects store")
	// ErrUnsupportedProtocol reports one unsupported requested Git protocol
	// version.
	ErrUnsupportedProtocol = errors.New("receivepack: unsupported protocol version")
)
