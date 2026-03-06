package commitquery

import "errors"

var errBadGenerationOrder = errors.New("commitquery: priority queue violated generation ordering")
