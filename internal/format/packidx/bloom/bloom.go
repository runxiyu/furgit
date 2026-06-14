package bloom

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"math/bits"

	"lindenii.org/go/furgit/object/id"
)

// ErrMalformedBloomFilter reports that
// a Bloom filter is truncated,
// has a bad signature, version, or hash function,
// or has inconsistent parameters.
var ErrMalformedBloomFilter = errors.New("internal/format/packidx/bloom: malformed bloom filter")

const (
	signature = 0x4944424c // "IDBL"
	version   = 1

	// HeaderLen is the fixed header length in octets,
	// i.e., the signature, version, hash function identifier,
	// B, K, and the trailing zero padding.
	HeaderLen = 64

	// BucketLen is the length of one bucket in octets,
	// chosen to match the most common cache-line size.
	BucketLen = 64

	// wordBits is the bit width of one bucket word.
	wordBits = 64

	// fieldBits is the width of one in-bucket position field.
	fieldBits = 9
)

// checkParams validates the filter parameters
// against one object hash size,
// returning log2(bucketCount) on success.
func checkParams(bucketCount uint32, k uint16, hashSize int) (uint, error) {
	switch {
	case bucketCount == 0 || bucketCount&(bucketCount-1) != 0:
		return 0, errors.New("bucket count not a nonzero power of two") //nolint:err113
	case k == 0:
		return 0, errors.New("zero probe count") //nolint:err113
	}

	log2B := uint(bits.TrailingZeros32(bucketCount))
	if log2B+fieldBits*uint(k) > uint(hashSize)*8 {
		return 0, errors.New("parameters exceed hash length") //nolint:err113
	}

	return log2B, nil
}

// hashFunctionID returns the on-disk hash function identifier
// for one object format.
func hashFunctionID(objectFormat id.ObjectFormat) (uint32, error) {
	switch objectFormat {
	case id.ObjectFormatSHA1:
		return 1, nil
	case id.ObjectFormatSHA256:
		return 2, nil
	case id.ObjectFormatUnknown:
	}

	return 0, id.ErrInvalidObjectFormat
}

// Bloom is a parsed blocked Bloom filter view over borrowed bytes.
//
// Labels: Deps-Borrowed, Life-Parent, MT-Safe.
type Bloom struct {
	// data is the entire filter payload.
	data []byte

	// buckets is the bucket region, between the header and the trailer.
	buckets []byte

	// objectFormat is the filter's object format.
	objectFormat id.ObjectFormat

	// log2B is the base-2 logarithm of the bucket count,
	// i.e. the number of leading object ID bits that select a bucket.
	log2B uint

	// k is the number of bits set and tested per object ID.
	k int
}

// Parse parses one Bloom filter from data.
//
// Labels: Deps-Borrowed, Life-Parent.
func Parse(data []byte, objectFormat id.ObjectFormat) (Bloom, error) {
	var zero Bloom

	wantHashID, err := hashFunctionID(objectFormat)
	if err != nil {
		return zero, err
	}

	hashSize := objectFormat.Size()

	if len(data) < HeaderLen {
		return zero, fmt.Errorf("%w: truncated", ErrMalformedBloomFilter)
	}

	if binary.BigEndian.Uint32(data) != signature {
		return zero, fmt.Errorf("%w: bad signature", ErrMalformedBloomFilter)
	}

	if binary.BigEndian.Uint32(data[4:]) != version {
		return zero, fmt.Errorf("%w: unsupported version", ErrMalformedBloomFilter)
	}

	if binary.BigEndian.Uint32(data[8:]) != wantHashID {
		return zero, fmt.Errorf("%w: hash function mismatch", ErrMalformedBloomFilter)
	}

	bucketCount := binary.BigEndian.Uint32(data[12:])
	k := binary.BigEndian.Uint16(data[16:])

	for _, octet := range data[18:HeaderLen] {
		if octet != 0 {
			return zero, fmt.Errorf("%w: nonzero padding", ErrMalformedBloomFilter)
		}
	}

	log2B, err := checkParams(bucketCount, k, hashSize)
	if err != nil {
		return zero, fmt.Errorf("%w: %w", ErrMalformedBloomFilter, err)
	}

	want := uint64(HeaderLen) + uint64(BucketLen)*uint64(bucketCount) + 2*uint64(hashSize)
	if uint64(len(data)) != want {
		return zero, fmt.Errorf("%w: file size disagrees with bucket count", ErrMalformedBloomFilter)
	}

	return Bloom{
		data:         data,
		buckets:      data[HeaderLen : len(data)-2*hashSize],
		objectFormat: objectFormat,
		log2B:        log2B,
		k:            int(k),
	}, nil
}

// PackHash returns the pack hash recorded in the filter trailer.
//
// Labels: Life-Parent, Mut-No.
func (f *Bloom) PackHash() []byte {
	hashSize := f.objectFormat.Size()
	end := len(f.data) - hashSize

	return f.data[end-hashSize : end]
}

// Verify recomputes the filter's trailing checksum and reports any mismatch.
//
// Verify reads the whole filter,
// so callers should treat it as a deliberate integrity check
// rather than part of the open path.
func (f *Bloom) Verify() error {
	hashImpl, err := f.objectFormat.New()
	if err != nil {
		return fmt.Errorf("internal/format/packidx/bloom: %w", err)
	}

	checksumOff := len(f.data) - f.objectFormat.Size()

	_, _ = hashImpl.Write(f.data[:checksumOff])

	if !bytes.Equal(hashImpl.Sum(nil), f.data[checksumOff:]) {
		return fmt.Errorf("%w: checksum mismatch", ErrMalformedBloomFilter)
	}

	return nil
}
