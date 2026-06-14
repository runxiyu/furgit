package main

import (
	"fmt"

	"lindenii.org/go/furgit/internal/cache/clock"
	"lindenii.org/go/furgit/internal/format/packfile"
	"lindenii.org/go/furgit/internal/format/packfile/delta"
	"lindenii.org/go/furgit/object/header"
	"lindenii.org/go/furgit/object/id"
	"lindenii.org/go/lgo/intconv"
)

const baseCacheMaxWeight = 64 << 20

type resolvedBase struct {
	entryType packfile.EntryType
	content   []byte
}

type baseCache = clock.Clock[int, resolvedBase]

func newBaseCache() *baseCache {
	return clock.New(baseCacheMaxWeight, func(_ int, base resolvedBase) uint64 {
		return uint64(len(base.content)) + 32
	})
}

func (explainer *explainer) reconstruct(offset, depth int) (packfile.EntryType, []byte, bool, error) {
	var zero packfile.EntryType

	if depth > delta.MaxChainDepth {
		return zero, nil, false, fmt.Errorf("delta chain too deep at offset %d", offset)
	}

	if cached, ok := explainer.cache.Get(offset); ok {
		return cached.entryType, cached.content, true, nil
	}

	header, err := packfile.ParseEntryHeader(explainer.data[offset:], explainer.objectFormat.Size())
	if err != nil {
		return zero, nil, false, fmt.Errorf("entry at offset %d: %w", offset, err)
	}

	payloadStart := offset + header.HeaderLen
	if payloadStart > len(explainer.data) {
		return zero, nil, false, fmt.Errorf("entry at offset %d: header runs past end of pack", offset)
	}

	if header.Type.IsBase() {
		content, _, err := inflateAt(explainer.data[payloadStart:])
		if err != nil {
			return zero, nil, false, fmt.Errorf("entry at offset %d: %w", offset, err)
		}

		explainer.cache.Add(offset, resolvedBase{entryType: header.Type, content: content})

		return header.Type, content, true, nil
	}

	baseOffset, ok, err := explainer.baseOffset(offset, header)
	if err != nil {
		return zero, nil, false, err
	}

	if !ok {
		return zero, nil, false, nil
	}

	baseType, baseContent, ok, err := explainer.reconstruct(baseOffset, depth+1)
	if err != nil || !ok {
		return zero, nil, ok, err
	}

	payload, _, err := inflateAt(explainer.data[payloadStart:])
	if err != nil {
		return zero, nil, false, fmt.Errorf("entry at offset %d: %w", offset, err)
	}

	content, err := delta.Apply(baseContent, payload)
	if err != nil {
		return zero, nil, false, fmt.Errorf("entry at offset %d: %w", offset, err)
	}

	explainer.cache.Add(offset, resolvedBase{entryType: baseType, content: content})

	return baseType, content, true, nil
}

func (explainer *explainer) baseOffset(offset int, header packfile.EntryHeader) (int, bool, error) {
	switch header.Type {
	case packfile.EntryTypeOfsDelta:
		dist, err := intconv.Uint64ToInt(header.OfsDistance)
		if err != nil || dist <= 0 || dist > offset {
			return 0, false, fmt.Errorf("entry at offset %d: ofs-delta base out of bounds", offset)
		}

		return offset - dist, true, nil
	case packfile.EntryTypeRefDelta:
		refBytes := header.RefBase[:explainer.objectFormat.Size()]

		if explainer.idx != nil {
			baseOffsetU, found, err := explainer.idx.Lookup(refBytes)
			if err != nil {
				return 0, false, fmt.Errorf("entry at offset %d: index lookup: %w", offset, err)
			}

			if found {
				baseOffset, err := intconv.Uint64ToInt(baseOffsetU)
				if err != nil {
					return 0, false, fmt.Errorf("entry at offset %d: index base offset overflows int: %w", offset, err)
				}

				return baseOffset, true, nil
			}
		}

		baseID, err := explainer.objectFormat.FromBytes(refBytes)
		if err != nil {
			return 0, false, fmt.Errorf("entry at offset %d: %w", offset, err)
		}

		if baseOffset, found := explainer.oidIndex[baseID]; found {
			return baseOffset, true, nil
		}

		return 0, false, nil
	case packfile.EntryTypeInvalid,
		packfile.EntryTypeCommit,
		packfile.EntryTypeTree,
		packfile.EntryTypeBlob,
		packfile.EntryTypeTag,
		packfile.EntryTypeFuture:
	}

	return 0, false, fmt.Errorf("entry at offset %d: not a delta entry", offset)
}

func (explainer *explainer) recomputeOID(entryType packfile.EntryType, content []byte) (id.ObjectID, error) {
	var zero id.ObjectID

	objectType, err := entryType.ObjectType()
	if err != nil {
		return zero, err
	}

	hashImpl, err := explainer.objectFormat.New()
	if err != nil {
		return zero, err
	}

	_, _ = hashImpl.Write(header.Append(nil, objectType, len(content)))
	_, _ = hashImpl.Write(content)

	oid, err := explainer.objectFormat.FromBytes(hashImpl.Sum(nil))
	if err != nil {
		return zero, err
	}

	return oid, nil
}
