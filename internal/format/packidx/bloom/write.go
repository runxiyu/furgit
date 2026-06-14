package bloom

import (
	"encoding/binary"
	"errors"
	"fmt"
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
	// data is the full filter file, header included.
	data []byte

	// buckets aliases the bucket region of data, after the header.
	buckets []byte

	log2B    uint
	k        int
	hashSize int
}

// NewBuilder creates a filter builder
// for bucketCount buckets and k probes per object ID.
//
// bucketCount must be a nonzero power of two,
// k must be nonzero,
// and log2(bucketCount) + 9*k must not exceed the hash length in bits.
func NewBuilder(objectFormat id.ObjectFormat, bucketCount uint32, k uint16) (*Builder, error) {
	hashID, err := hashFunctionID(objectFormat)
	if err != nil {
		return nil, err
	}

	hashSize := objectFormat.Size()

	log2B, err := checkParams(bucketCount, k, hashSize)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrInvalidParameters, err)
	}

	total, err := intconv.Uint64ToInt(uint64(HeaderLen) + uint64(BucketLen)*uint64(bucketCount))
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrInvalidParameters, err)
	}

	data := make([]byte, total)
	binary.BigEndian.PutUint32(data[0:], signature)
	binary.BigEndian.PutUint32(data[4:], version)
	binary.BigEndian.PutUint32(data[8:], hashID)
	binary.BigEndian.PutUint32(data[12:], bucketCount)
	binary.BigEndian.PutUint16(data[16:], k)

	return &Builder{
		data:     data,
		buckets:  data[HeaderLen:],
		log2B:    log2B,
		k:        int(k),
		hashSize: hashSize,
	}, nil
}

// Add records oid in the filter.
//
// oid must be exactly the filter's hash size;
// Add panics otherwise.
func (b *Builder) Add(oid []byte) {
	if len(oid) != b.hashSize {
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

// Bytes returns the serialized filter.
//
// Labels: Life-Parent, Mut-No.
func (b *Builder) Bytes() []byte {
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
