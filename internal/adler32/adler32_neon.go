//go:build !purego && arm64

package adler32

//go:noescape
func adler32_neon(in uint32, buf []byte) uint32
