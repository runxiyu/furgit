package bloom

import "codeberg.org/lindenii/furgit/internal/intconv"

func murmur3SeededV2(seed uint32, data []byte) (uint32, error) {
	const (
		c1 = 0xcc9e2d51
		c2 = 0x1b873593
		r1 = 15
		r2 = 13
		m  = 5
		n  = 0xe6546b64
	)

	h := seed

	nblocks := len(data) / 4
	for i := range nblocks {
		k := uint32(data[4*i]) |
			(uint32(data[4*i+1]) << 8) |
			(uint32(data[4*i+2]) << 16) |
			(uint32(data[4*i+3]) << 24)
		k *= c1
		k = (k << r1) | (k >> (32 - r1))
		k *= c2

		h ^= k
		h = (h << r2) | (h >> (32 - r2))
		h = h*m + n
	}

	var k1 uint32

	tail := data[nblocks*4:]
	switch len(tail) & 3 {
	case 3:
		k1 ^= uint32(tail[2]) << 16

		fallthrough
	case 2:
		k1 ^= uint32(tail[1]) << 8

		fallthrough
	case 1:
		k1 ^= uint32(tail[0])
		k1 *= c1
		k1 = (k1 << r1) | (k1 >> (32 - r1))
		k1 *= c2
		h ^= k1
	}

	dataLen, err := intconv.IntToUint32(len(data))
	if err != nil {
		return 0, err
	}

	h ^= dataLen
	h ^= h >> 16
	h *= 0x85ebca6b
	h ^= h >> 13
	h *= 0xc2b2ae35
	h ^= h >> 16

	return h, nil
}

func murmur3SeededV1(seed uint32, data []byte) (uint32, error) {
	const (
		c1 = 0xcc9e2d51
		c2 = 0x1b873593
		r1 = 15
		r2 = 13
		m  = 5
		n  = 0xe6546b64
	)

	h := seed

	nblocks := len(data) / 4
	for i := range nblocks {
		k := intconv.SignExtendByteToUint32(data[4*i]) |
			(intconv.SignExtendByteToUint32(data[4*i+1]) << 8) |
			(intconv.SignExtendByteToUint32(data[4*i+2]) << 16) |
			(intconv.SignExtendByteToUint32(data[4*i+3]) << 24)
		k *= c1
		k = (k << r1) | (k >> (32 - r1))
		k *= c2

		h ^= k
		h = (h << r2) | (h >> (32 - r2))
		h = h*m + n
	}

	var k1 uint32

	tail := data[nblocks*4:]
	switch len(tail) & 3 {
	case 3:
		k1 ^= intconv.SignExtendByteToUint32(tail[2]) << 16

		fallthrough
	case 2:
		k1 ^= intconv.SignExtendByteToUint32(tail[1]) << 8

		fallthrough
	case 1:
		k1 ^= intconv.SignExtendByteToUint32(tail[0])
		k1 *= c1
		k1 = (k1 << r1) | (k1 >> (32 - r1))
		k1 *= c2
		h ^= k1
	}

	dataLen, err := intconv.IntToUint32(len(data))
	if err != nil {
		return 0, err
	}

	h ^= dataLen
	h ^= h >> 16
	h *= 0x85ebca6b
	h ^= h >> 13
	h *= 0xc2b2ae35
	h ^= h >> 16

	return h, nil
}
