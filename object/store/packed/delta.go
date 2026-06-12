package packed

import (
	"bytes"
	"fmt"
	"io"
	"slices"

	"lindenii.org/go/furgit/errs"
	"lindenii.org/go/furgit/internal/compress/zlib"
	"lindenii.org/go/furgit/internal/format/packfile"
	"lindenii.org/go/furgit/internal/format/packfile/delta"
)

// maxDeltaDepth bounds delta chain length,
// guaranteeing termination on crafted ref-delta loops.
const maxDeltaDepth = 1 << 12

// deltaNode is a delta entry on a resolution chain.
type deltaNode struct {
	// payload is the entry's compressed delta payload view.
	payload []byte

	// size is the entry's declared inflated delta size.
	size uint64

	// baseOffset is the entry's base entry offset.
	baseOffset uint64
}

// unpackEntry reconstructs the object stored at offset in p,
// following ref- and ofs-delta chains within the pack.
//
// Labels: Life-Independent.
func (packed *Packed) unpackEntry(p *pack, offset uint64) (packfile.EntryType, []byte, error) {
	var zero packfile.EntryType

	var (
		chain     []deltaNode
		baseType  packfile.EntryType
		base      []byte
		fromCache bool
	)

	// Drill down to the innermost base,
	// stopping early at any cached base.
	cur := offset

	for {
		if cached, ok := packed.baseCache.Get(baseKey{pack: p, offset: cur}); ok {
			baseType = cached.entryType
			base = cached.content
			fromCache = true

			break
		}

		if len(chain) >= maxDeltaDepth {
			return zero, nil, fmt.Errorf("%w: pack %q: delta chain too deep", ErrMalformedPackedStore, p.name)
		}

		header, payload, err := p.entryHeaderAt(cur, packed.objectFormat)
		if err != nil {
			return zero, nil, err
		}

		if header.Type.IsBase() {
			base, err = inflate(payload, header.Size)
			if err != nil {
				return zero, nil, fmt.Errorf("%w: pack %q: %w", ErrMalformedPackedStore, p.name, err)
			}

			baseType = header.Type

			break
		}

		baseOffset, err := packed.deltaBaseOffset(p, cur, header)
		if err != nil {
			return zero, nil, err
		}

		chain = append(chain, deltaNode{
			payload:    payload,
			size:       header.Size,
			baseOffset: baseOffset,
		})

		cur = baseOffset
	}

	// A direct cache hit with no deltas to apply must be copied.
	if len(chain) == 0 && fromCache {
		return baseType, bytes.Clone(base), nil
	}

	// Apply deltas back up the chain, caching each consumed base.
	for i, node := range slices.Backward(chain) {
		deltaData, err := inflate(node.payload, node.size)
		if err != nil {
			return zero, nil, fmt.Errorf("%w: pack %q: %w", ErrMalformedPackedStore, p.name, err)
		}

		result, err := delta.Apply(base, deltaData)
		if err != nil {
			return zero, nil, fmt.Errorf("%w: pack %q: %w", ErrMalformedPackedStore, p.name, err)
		}

		consumedFromCache := fromCache && i == len(chain)-1
		if !consumedFromCache {
			packed.baseCache.Add(
				baseKey{pack: p, offset: node.baseOffset},
				cachedBase{entryType: baseType, content: base},
			)
		}

		base = result
	}

	return baseType, base, nil
}

// deltaBaseOffset resolves a delta entry's base entry offset
// within the same pack.
func (packed *Packed) deltaBaseOffset(p *pack, offset uint64, header packfile.EntryHeader) (uint64, error) {
	switch header.Type {
	case packfile.EntryTypeOfsDelta:
		if header.OfsDistance == 0 || header.OfsDistance > offset {
			return 0, fmt.Errorf("%w: pack %q: invalid ofs-delta distance", ErrMalformedPackedStore, p.name)
		}

		return offset - header.OfsDistance, nil
	case packfile.EntryTypeRefDelta:
		refBase := header.RefBase[:packed.objectFormat.Size()]

		baseOffset, found, err := p.idx.Lookup(refBase)
		if err != nil {
			return 0, fmt.Errorf("%w: pack %q: %w", ErrMalformedPackedStore, p.name, err)
		}

		if !found {
			baseID, idErr := packed.objectFormat.FromBytes(refBase)
			if idErr != nil {
				return 0, fmt.Errorf("%w: pack %q: %w", ErrMalformedPackedStore, p.name, idErr)
			}

			return 0, fmt.Errorf(
				"%w: resolving ref-delta: %w",
				ErrMalformedPackedStore, &errs.ObjectMissingError{OID: baseID},
			)
		}

		return baseOffset, nil
	case packfile.EntryTypeInvalid,
		packfile.EntryTypeCommit,
		packfile.EntryTypeTree,
		packfile.EntryTypeBlob,
		packfile.EntryTypeTag,
		packfile.EntryTypeFuture:
	}

	panic("object/store/packed: deltaBaseOffset on non-delta entry")
}

// resolveType walks one delta chain
// to find the chained base object entry type,
// without inflating any content.
func (packed *Packed) resolveType(p *pack, offset uint64, entryHeader packfile.EntryHeader) (packfile.EntryType, error) {
	var zero packfile.EntryType

	depth := 0

	for entryHeader.Type.IsDelta() {
		if cached, ok := packed.baseCache.Peek(baseKey{pack: p, offset: offset}); ok {
			return cached.entryType, nil
		}

		depth++
		if depth > maxDeltaDepth {
			return zero, fmt.Errorf("%w: pack %q: delta chain too deep", ErrMalformedPackedStore, p.name)
		}

		baseOffset, err := packed.deltaBaseOffset(p, offset, entryHeader)
		if err != nil {
			return zero, err
		}

		offset = baseOffset

		entryHeader, _, err = p.entryHeaderAt(offset, packed.objectFormat)
		if err != nil {
			return zero, err
		}
	}

	return entryHeader.Type, nil
}

// deltaResultSize reads the declared result size
// from one compressed delta payload prefix.
func deltaResultSize(payload []byte, deltaSize uint64) (uint64, error) {
	zr, err := zlib.NewReader(bytes.NewReader(payload))
	if err != nil {
		return 0, fmt.Errorf("reading delta header: %w", err)
	}

	defer func() { _ = zr.Close() }()

	prefixLen := min(uint64(delta.MaxHeaderSizesLen), deltaSize)

	prefix := make([]byte, prefixLen)

	_, err = io.ReadFull(zr, prefix)
	if err != nil {
		return 0, fmt.Errorf("reading delta header: %w", err)
	}

	_, resultSize, _, err := delta.ParseHeaderSizes(prefix)
	if err != nil {
		return 0, fmt.Errorf("reading delta header: %w", err)
	}

	return resultSize, nil
}
