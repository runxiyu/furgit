package commitquery

import "errors"

// errBadGenerationOrder reports an invalid priority-queue ordering.
var errBadGenerationOrder = errors.New("commitquery: priority queue violated generation ordering")
