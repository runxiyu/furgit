// Package bloom provides a blocked Bloom filter
// for pack indexes.
//
// A filter answers, from a single cache-line-sized read,
// whether an object ID is definitely absent from the index it covers.
// A lookup that must consult many packs
// can then skip the full binary search
// in every pack whose filter rejects the object,
// decreasing the cost of misses.
//
// # Rationale
//
// Especially for server-side usage,
// repacking is expensive,
// and creating multi-pack-indexes is still rather expensive.
// Incremental multi-pack-indexes partially solve this,
// but having too many of them defeats the purpose,
// since the indexes must still be walked in order
// while performing expensive lookups.
//
// Instead, each multi-pack-index layer,
// and each ordinary pack index,
// may carry its own filter.
// The indexes are still traversed in their usual order,
// but the first step when traversing one
// is to check whether it could possibly hold the wanted object.
//
// The filter is split into 64-octet buckets,
// matching the most common cache-line size.
// Some bits of the object ID choose the bucket,
// and the rest choose several bit positions inside it,
// so a lookup reads one 64-octet bucket
// and checks whether all required bits are set.
//
// # Parameters
//
// A filter is parameterized by
// the number of buckets B
// and the number of bits set and tested per object ID, K.
// All integers in the format are big endian.
// The object ID is interpreted as a big-endian bitstring,
// where bit offset 0 is the most significant bit of octet 0.
// B must be a nonzero power of two,
// K must be nonzero,
// and log2(B) + 9*K must not exceed the hash length in bits.
//
// # File format
//
// A filter file is a 64-octet header
// followed by B buckets of 64 octets each:
//
//   - 4-octet signature: {'I', 'D', 'B', 'L'}.
//   - 4-octet version identifier (= 1).
//   - 4-octet object hash algorithm identifier
//     (= 1 for SHA-1, 2 for SHA-256).
//   - 4-octet B, the number of buckets.
//   - 2-octet K, the number of bits set and tested per object ID.
//   - 46-octet padding, which must be all zero.
//   - B buckets of 64 octets each.
//
// A reader must validate that
// the signature matches,
// the version is supported,
// the hash function identifier is recognized,
// B is nonzero and a power of two,
// K is nonzero,
// log2(B) + 9*K does not exceed the hash length in bits,
// the padding is all zero,
// and the file size is exactly 64 + 64*B octets.
//
// # Lookup
//
// A lookup against one filter proceeds as follows:
//
//  1. Let b be the unsigned integer encoded
//     by the most significant log2(B) bits of the object ID.
//     B is a power of two, so 0 <= b < B.
//  2. Select and read bucket b.
//  3. For each 0 <= i < K,
//     take the i-th 9-bit field
//     from the 9*K bits that follow the bucket-selecting bits,
//     and let pi be the unsigned integer it encodes,
//     so 0 <= pi < 512.
//     Compute wi = pi >> 6 and bi = pi & 63,
//     so wi identifies one of the eight 64-bit words in bucket b
//     and bi identifies one bit within that word.
//     Within each 64-bit word,
//     bit index 0 is the most significant bit
//     and bit index 63 is the least significant bit.
//     Test whether bit bi is set in word wi of bucket b.
//
// If any test fails,
// the object ID is definitely not in the covered index.
// If all tests succeed,
// the object ID may be in it.
// Two of the K 9-bit fields can decode to the same pi,
// so an insertion may set fewer than K distinct bits;
// this only raises the false positive rate
// and never causes a false negative.
//
// # Worked example
//
// Let B = 1 << 15 = 32768 and K = 8.
// Then log2(B) = 15,
// so each lookup uses 15 bits to choose the bucket
// and 8*9 = 72 bits to choose the in-bucket positions,
// for a total of 87 bits taken from the object ID.
// A SHA-1 has 160 bits and a SHA-256 has 256 bits,
// so both leave ample headroom.
//
// # Security considerations
//
// Object IDs are public unkeyed hashes,
// so an adversary can mine packs
// whose object IDs share a chosen prefix
// to crowd objects into one bucket
// and fill its bits.
// In the worst case this renders some buckets useless,
// making the filter degrade to "may contain" for those buckets,
// but it never produces a false negative
// and is not a significant denial-of-service vector.
package bloom
