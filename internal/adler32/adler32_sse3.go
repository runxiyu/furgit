//go:build !purego && amd64

package adler32

//go:noescape
func adler32_sse3(in uint32, buf []byte) uint32
