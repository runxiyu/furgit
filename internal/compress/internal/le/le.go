// Package le provides fast little endian integer routines.
package le

type Indexer interface {
	int | int8 | int16 | int32 | int64 | uint | uint8 | uint16 | uint32 | uint64
}
