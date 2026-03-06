package bufpool

//nolint:gochecknoglobals
var sizeClasses = [...]int{
	DefaultBufferCap,
	64 << 10,
	128 << 10,
	256 << 10,
	512 << 10,
	1 << 20,
	2 << 20,
	4 << 20,
	maxPooledBuffer,
}

func classFor(size int) (idx, classCap int, ok bool) {
	for i, class := range sizeClasses {
		if size <= class {
			return i, class, true
		}
	}

	return -1, size, false
}
