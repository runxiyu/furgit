package bloom

import (
	"encoding/binary"
)

// MayContain reports whether oid may be present
// in the index covered by the filter.
//
// oid must be exactly the filter's hash size;
// MayContain panics otherwise.
//
// Labels: Mut-No.
func (f *Bloom) MayContain(oid []byte) bool {
	if len(oid) != f.objectFormat.Size() {
		panic("internal/format/packidx/bloom: invalid object ID length")
	}

	base := int(binary.BigEndian.Uint32(oid[:4])>>(32-f.log2B)) * BucketLen

	for i := range f.k {
		word, mask := probe(oid, f.log2B, i)
		if binary.BigEndian.Uint64(f.buckets[base+word*8:])&mask == 0 {
			return false
		}
	}

	return true
}

// probe returns the bucket word index and single-bit mask
// addressed by the i-th probe of oid.
func probe(oid []byte, log2B uint, i int) (word int, mask uint64) {
	bitOff := log2B + fieldBits*uint(i)
	byteOff := bitOff >> 3
	bitInByte := bitOff & 7

	window := uint32(oid[byteOff])<<8 | uint32(oid[byteOff+1])
	pi := (window >> (16 - bitInByte - fieldBits)) & 0x1ff

	return int(pi >> 6), 1 << (wordBits - 1 - (pi & 63))
}
