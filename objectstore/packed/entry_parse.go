package packed

import (
	"fmt"

	"codeberg.org/lindenii/furgit/objectid"
	"codeberg.org/lindenii/furgit/objecttype"
)

// entryMeta describes one parsed pack entry header.
type entryMeta struct {
	// ty is the pack entry type tag.
	ty objecttype.Type
	// size is the declared resulting content size.
	size int64
	// dataOffset points to the zlib payload start.
	dataOffset int
	// baseRefID is set for ref-delta entries.
	baseRefID objectid.ObjectID
	// baseOfs is set for ofs-delta entries.
	baseOfs uint64
}

// parseEntryMeta parses one pack entry header at offset.
func parseEntryMeta(pack *packFile, algo objectid.Algorithm, offset uint64) (entryMeta, error) {
	var zero entryMeta
	if offset >= uint64(len(pack.data)) {
		return zero, fmt.Errorf("objectstore/packed: pack %q offset %d out of bounds", pack.name, offset)
	}

	pos := int(offset)
	first := pack.data[pos]
	pos++

	meta := entryMeta{
		ty:   objecttype.Type((first >> 4) & 0x07),
		size: int64(first & 0x0f),
	}

	shift := uint(4)
	b := first
	for b&0x80 != 0 {
		if pos >= len(pack.data) {
			return zero, fmt.Errorf("objectstore/packed: pack %q truncated entry header", pack.name)
		}
		b = pack.data[pos]
		pos++
		meta.size |= int64(b&0x7f) << shift
		shift += 7
	}
	if meta.size < 0 {
		return zero, fmt.Errorf("objectstore/packed: pack %q entry has negative size", pack.name)
	}

	switch meta.ty {
	case objecttype.TypeCommit, objecttype.TypeTree, objecttype.TypeBlob, objecttype.TypeTag:
		// Base object entries have no extra header fields.
	case objecttype.TypeRefDelta:
		hashSize := algo.Size()
		if pos+hashSize > len(pack.data) {
			return zero, fmt.Errorf("objectstore/packed: pack %q truncated ref-delta base id", pack.name)
		}
		baseID, err := objectid.FromBytes(algo, pack.data[pos:pos+hashSize])
		if err != nil {
			return zero, err
		}
		meta.baseRefID = baseID
		pos += hashSize
	case objecttype.TypeOfsDelta:
		dist, consumed, err := parseOfsDeltaDistance(pack.data[pos:])
		if err != nil {
			return zero, err
		}
		pos += consumed
		if offset <= dist {
			return zero, fmt.Errorf("objectstore/packed: pack %q has invalid ofs-delta base", pack.name)
		}
		meta.baseOfs = offset - dist
	default:
		return zero, fmt.Errorf("objectstore/packed: pack %q has unsupported object type %d", pack.name, meta.ty)
	}

	meta.dataOffset = pos
	if meta.dataOffset > len(pack.data) {
		return zero, fmt.Errorf("objectstore/packed: pack %q entry data offset out of bounds", pack.name)
	}
	return meta, nil
}

// parseOfsDeltaDistance parses one ofs-delta backward distance.
func parseOfsDeltaDistance(buf []byte) (uint64, int, error) {
	if len(buf) == 0 {
		return 0, 0, fmt.Errorf("objectstore/packed: malformed ofs-delta distance")
	}
	b := buf[0]
	dist := uint64(b & 0x7f)
	consumed := 1
	for b&0x80 != 0 {
		if consumed >= len(buf) {
			return 0, 0, fmt.Errorf("objectstore/packed: malformed ofs-delta distance")
		}
		b = buf[consumed]
		consumed++
		dist = ((dist + 1) << 7) + uint64(b&0x7f)
	}
	return dist, consumed, nil
}

// isBaseObjectType reports whether ty is one of the four canonical object types.
func isBaseObjectType(ty objecttype.Type) bool {
	switch ty {
	case objecttype.TypeCommit, objecttype.TypeTree, objecttype.TypeBlob, objecttype.TypeTag:
		return true
	default:
		return false
	}
}
