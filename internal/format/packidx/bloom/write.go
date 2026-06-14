package bloom

import (
	"encoding/binary"
	"errors"
	"fmt"
	"hash"
	"math/bits"

	"lindenii.org/go/furgit/object/id"
	"lindenii.org/go/lgo/intconv"
)

// ErrInvalidParameters reports that
// the parameters supplied for a filter build
// are not representable in the format.
var ErrInvalidParameters = errors.New("internal/format/packidx/bloom: invalid parameters")

// defaultK is the probe count used by [RecommendParams].
//
// With 512-bit buckets it keeps the false positive rate near one percent
// at the target bucket load.
const defaultK = 8

// targetLoad is the object count per bucket that [RecommendParams] aims for.
const targetLoad = 48

// Builder accumulates object IDs into an in-memory Bloom filter
// and serializes it.
//
// Labels: MT-Unsafe.
type Builder struct {
	// data is the full filter file, header and trailer included.
	data []byte

	// buckets aliases the bucket region of data, between header and trailer.
	buckets []byte

	// hashImpl computes the trailing checksum and gives the hash size.
	hashImpl hash.Hash

	log2B uint
	k     int
}

// NewBuilder creates a filter builder
// for bucketCount buckets and k probes per object ID,
// binding the filter to packHash.
//
// bucketCount must be a nonzero power of two,
// k must be nonzero,
// and log2(bucketCount) + 9*k must not exceed the hash length in bits.
// packHash must be the pack's trailer hash;
// NewBuilder panics when its length does not match the object format.
func NewBuilder(objectFormat id.ObjectFormat, bucketCount uint32, k uint16, packHash []byte) (*Builder, error) {
	hashID, err := hashFunctionID(objectFormat)
	if err != nil {
		return nil, err
	}

	hashImpl, err := objectFormat.New()
	if err != nil {
		return nil, fmt.Errorf("internal/format/packidx/bloom: %w", err)
	}

	hashSize := objectFormat.Size()

	if len(packHash) != hashSize {
		panic("internal/format/packidx/bloom: invalid pack hash length")
	}

	log2B, err := checkParams(bucketCount, k, hashSize)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrInvalidParameters, err)
	}

	total, err := intconv.Uint64ToInt(uint64(HeaderLen) + uint64(BucketLen)*uint64(bucketCount) + 2*uint64(hashSize))
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrInvalidParameters, err)
	}

	data := make([]byte, total)
	binary.BigEndian.PutUint32(data[0:], signature)
	binary.BigEndian.PutUint32(data[4:], version)
	binary.BigEndian.PutUint32(data[8:], hashID)
	binary.BigEndian.PutUint32(data[12:], bucketCount)
	binary.BigEndian.PutUint16(data[16:], k)

	bucketsEnd := total - 2*hashSize
	copy(data[bucketsEnd:], packHash)

	return &Builder{
		data:     data,
		buckets:  data[HeaderLen:bucketsEnd],
		hashImpl: hashImpl,
		log2B:    log2B,
		k:        int(k),
	}, nil
}

// Add records oid in the filter.
//
// oid must be exactly the filter's hash size;
// Add panics otherwise.
func (b *Builder) Add(oid []byte) {
	if len(oid) != b.hashImpl.Size() {
		panic("internal/format/packidx/bloom: invalid object ID length")
	}

	base := int(binary.BigEndian.Uint32(oid[:4])>>(32-b.log2B)) * BucketLen

	for i := range b.k {
		word, mask := probe(oid, b.log2B, i)

		off := base + word*8
		set := binary.BigEndian.Uint64(b.buckets[off:]) | mask
		binary.BigEndian.PutUint64(b.buckets[off:], set)
	}
}

// Bytes returns the serialized filter, including its trailing checksum.
//
// Labels: Life-Parent, Mut-No.
func (b *Builder) Bytes() []byte {
	checksumOff := len(b.data) - b.hashImpl.Size()

	b.hashImpl.Reset()
	_, _ = b.hashImpl.Write(b.data[:checksumOff])
	b.hashImpl.Sum(b.data[checksumOff:checksumOff])

	return b.data
}

// RecommendParams returns filter parameters for an index of n objects,
// targeting a false positive rate near one percent.
func RecommendParams(objectFormat id.ObjectFormat, n int) (bucketCount uint32, k uint16, err error) {
	hashSize := objectFormat.Size()
	if hashSize == 0 {
		return 0, 0, id.ErrInvalidObjectFormat
	}

	const maxPow2 = uint32(1) << 31

	wanted := uint64(0)
	if n > 0 {
		wanted = (uint64(n) + targetLoad - 1) / targetLoad
	}

	switch {
	case wanted <= 1:
		bucketCount = 1
	case wanted > uint64(maxPow2):
		bucketCount = maxPow2
	default:
		bucketCount = uint32(1) << bits.Len64(wanted-1)
	}

	_, err = checkParams(bucketCount, defaultK, hashSize)
	if err != nil {
		return 0, 0, fmt.Errorf("%w: %w", ErrInvalidParameters, err)
	}

	return bucketCount, defaultK, nil
}
