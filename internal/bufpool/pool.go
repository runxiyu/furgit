package bufpool

import "sync"

var bufferPools = func() []sync.Pool {
	pools := make([]sync.Pool, len(sizeClasses))
	for i, classCap := range sizeClasses {
		capCopy := classCap
		pools[i].New = func() any {
			buf := make([]byte, 0, capCopy)

			return &buf
		}
	}

	return pools
}()
