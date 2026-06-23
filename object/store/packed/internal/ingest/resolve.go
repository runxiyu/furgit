package ingest

import (
	"fmt"
	"io"

	"lindenii.org/go/furgit/internal/compress/zlib"
	"lindenii.org/go/furgit/internal/format/packfile"
	"lindenii.org/go/furgit/internal/format/packfile/delta"
	"lindenii.org/go/furgit/internal/progress"
	"lindenii.org/go/furgit/object/header"
	"lindenii.org/go/furgit/object/id"
	"lindenii.org/go/furgit/object/store"
	"lindenii.org/go/furgit/object/typ"
)

// adjacency maps each resolvable base to its delta children:
// ofs-deltas keyed by base offset, ref-deltas keyed by base object ID.
type adjacency struct {
	byOffset map[int][]int
	byOID    map[id.ObjectID][]int
}

// resolveDeltas resolves every delta record into a final object ID and type,
// completing thin packs from the external base reader when required.
func (ingestion *ingestion) resolveDeltas() error {
	meter := progress.New(progress.Options{
		Writer:     ingestion.opts.Progress,
		Title:      "resolving deltas",
		Total:      ingestion.countDeltas(),
		Delay:      0,
		Sparse:     false,
		Throughput: false,
	})

	adjacency := ingestion.buildAdjacency()

	err := ingestion.resolveFrom(ingestion.resolvedRoots(), adjacency, meter)
	if err != nil {
		return err
	}

	external := ingestion.unresolvedExternalBases()

	switch {
	case len(external) == 0 && ingestion.countUnresolved() > 0:
		return fmt.Errorf("%w: unresolvable delta entries", ErrMalformedPack)
	case len(external) > 0:
		err = ingestion.fixThin(external, adjacency, meter)
		if err != nil {
			return err
		}
	}

	meter.Stop("done")

	return nil
}

// buildAdjacency indexes every delta record by its base,
// so a resolved base can find the children that delta against it.
func (ingestion *ingestion) buildAdjacency() adjacency {
	out := adjacency{
		byOffset: make(map[int][]int),
		byOID:    make(map[id.ObjectID][]int),
	}

	for index := range ingestion.records {
		rec := &ingestion.records[index]

		switch rec.packedType {
		case packfile.EntryTypeOfsDelta:
			out.byOffset[rec.baseOffset] = append(out.byOffset[rec.baseOffset], index)
		case packfile.EntryTypeRefDelta:
			out.byOID[rec.baseOID] = append(out.byOID[rec.baseOID], index)
		case packfile.EntryTypeInvalid,
			packfile.EntryTypeCommit,
			packfile.EntryTypeTree,
			packfile.EntryTypeBlob,
			packfile.EntryTypeTag,
			packfile.EntryTypeFuture:
		}
	}

	return out
}

// resolveFrame is a resolved record whose delta children remain to be resolved.
type resolveFrame struct {
	index int
	depth int
}

func (ingestion *ingestion) resolveFrom(roots []int, adjacency adjacency, meter *progress.Meter) error {
	stack := make([]resolveFrame, 0, len(roots))
	for _, root := range roots {
		stack = append(stack, resolveFrame{index: root, depth: 0})
	}

	for len(stack) > 0 {
		err := ingestion.ctx.Err()
		if err != nil {
			return fmt.Errorf("object/store/packed/internal/ingest: %w", err)
		}

		frame := stack[len(stack)-1]
		stack = stack[:len(stack)-1]

		rec := &ingestion.records[frame.index]

		children := [2][]int{adjacency.byOffset[rec.offset], adjacency.byOID[rec.oid]}
		if len(children[0]) == 0 && len(children[1]) == 0 {
			continue
		}

		baseType, baseContent, err := ingestion.materialize(frame.index)
		if err != nil {
			return err
		}

		childDepth := frame.depth + 1

		for _, group := range children {
			for _, child := range group {
				if ingestion.records[child].resolved {
					continue
				}

				if childDepth > delta.MaxChainDepth {
					return fmt.Errorf("%w: entry at %d: delta chain too deep", ErrMalformedPack, ingestion.records[child].offset)
				}

				err = ingestion.resolveOneChild(child, baseType, baseContent, meter)
				if err != nil {
					return err
				}

				stack = append(stack, resolveFrame{index: child, depth: childDepth})
			}
		}
	}

	return nil
}

func (ingestion *ingestion) resolveOneChild(index int, baseType typ.Type, baseContent []byte, meter *progress.Meter) error {
	rec := &ingestion.records[index]

	content, err := ingestion.applyDelta(index, baseContent)
	if err != nil {
		return err
	}

	oid, err := ingestion.hashObject(baseType, content)
	if err != nil {
		return err
	}

	rec.oid = oid
	rec.resolved = true
	ingestion.byOID[oid] = index
	ingestion.baseCache.Add(baseCacheKey{offset: rec.offset}, cachedContent{objectType: baseType, content: content})

	ingestion.deltasResolved++
	meter.Set(ingestion.deltasResolved, 0)

	return nil
}

// materialize returns the inflated content of an already-resolved record,
// from the base cache,
// or re-derived from the nearest cached or base ancestor on a miss.
func (ingestion *ingestion) materialize(index int) (typ.Type, []byte, error) {
	var (
		zero     typ.Type
		chain    []int
		base     []byte
		baseType typ.Type
	)

	cur := index

	for {
		rec := &ingestion.records[cur]

		if cached, ok := ingestion.baseCache.Get(baseCacheKey{offset: rec.offset}); ok {
			base = cached.content
			baseType = cached.objectType

			break
		}

		if rec.packedType.IsBase() {
			objectType, err := rec.packedType.ObjectType()
			if err != nil {
				return zero, nil, fmt.Errorf("object/store/packed/internal/ingest: %w", err)
			}

			content, err := ingestion.inflateRecord(cur)
			if err != nil {
				return zero, nil, err
			}

			base = content
			baseType = objectType

			break
		}

		if len(chain) >= delta.MaxChainDepth {
			return zero, nil, fmt.Errorf("%w: entry at %d: delta chain too deep", ErrMalformedPack, rec.offset)
		}

		chain = append(chain, cur)

		next, ok := ingestion.baseRecordIndex(rec)
		if !ok {
			return zero, nil, fmt.Errorf("%w: entry at %d: base unavailable while reconstructing", ErrMalformedPack, rec.offset)
		}

		cur = next
	}

	for i := len(chain) - 1; i >= 0; i-- {
		content, err := ingestion.applyDelta(chain[i], base)
		if err != nil {
			return zero, nil, err
		}

		ingestion.baseCache.Add(baseCacheKey{offset: ingestion.records[chain[i]].offset}, cachedContent{objectType: baseType, content: content})

		base = content
	}

	return baseType, base, nil
}

func (ingestion *ingestion) applyDelta(index int, baseContent []byte) ([]byte, error) {
	rec := &ingestion.records[index]

	deltaPayload, err := ingestion.inflateRecord(index)
	if err != nil {
		return nil, err
	}

	baseSize, resultSize, _, err := delta.ParseHeaderSizes(deltaPayload)
	if err != nil {
		return nil, fmt.Errorf("%w: entry at %d: %w", ErrMalformedPack, rec.offset, err)
	}

	limit := ingestion.opts.MaxObjectSize
	if limit > 0 && resultSize > uint64(limit) {
		return nil, fmt.Errorf("%w: entry at %d: result size %d exceeds limit %d", store.ErrObjectTooLarge, rec.offset, resultSize, limit)
	}

	if baseSize != uint64(len(baseContent)) {
		return nil, fmt.Errorf("%w: entry at %d: delta base size mismatch", ErrMalformedPack, rec.offset)
	}

	content, err := delta.Apply(baseContent, deltaPayload)
	if err != nil {
		return nil, fmt.Errorf("%w: entry at %d: %w", ErrMalformedPack, rec.offset, err)
	}

	if uint64(len(content)) != resultSize {
		return nil, fmt.Errorf("%w: entry at %d: delta result size mismatch", ErrMalformedPack, rec.offset)
	}

	return content, nil
}

func (ingestion *ingestion) baseRecordIndex(rec *record) (int, bool) {
	switch rec.packedType {
	case packfile.EntryTypeOfsDelta:
		index, ok := ingestion.byOffset[rec.baseOffset]

		return index, ok
	case packfile.EntryTypeRefDelta:
		index, ok := ingestion.byOID[rec.baseOID]

		return index, ok
	case packfile.EntryTypeInvalid,
		packfile.EntryTypeCommit,
		packfile.EntryTypeTree,
		packfile.EntryTypeBlob,
		packfile.EntryTypeTag,
		packfile.EntryTypeFuture:
	}

	return 0, false
}

// inflateRecord inflates one record's payload from the temporary pack file.
func (ingestion *ingestion) inflateRecord(index int) ([]byte, error) {
	rec := &ingestion.records[index]

	offset := int64(rec.dataOffset())
	compressedLen := int64(rec.packedLen - rec.headerLen)
	size := rec.declaredSize

	zr, err := zlib.NewReader(io.NewSectionReader(ingestion.packFile, offset, compressedLen))
	if err != nil {
		return nil, fmt.Errorf("%w: entry at %d: %w", ErrMalformedPack, rec.offset, err)
	}

	defer func() { _ = zr.Close() }()

	out := make([]byte, size)

	_, err = io.ReadFull(zr, out)
	if err != nil {
		return nil, fmt.Errorf("%w: entry at %d: %w", ErrMalformedPack, rec.offset, err)
	}

	return out, nil
}

// hashObject computes the object ID of one resolved object.
func (ingestion *ingestion) hashObject(objectType typ.Type, content []byte) (id.ObjectID, error) {
	var zero id.ObjectID

	hashImpl, err := ingestion.objectFormat.New()
	if err != nil {
		return zero, fmt.Errorf("object/store/packed/internal/ingest: %w", err)
	}

	_, _ = hashImpl.Write(header.Append(nil, objectType, len(content)))
	_, _ = hashImpl.Write(content)

	oid, err := ingestion.objectFormat.FromBytes(hashImpl.Sum(nil))
	if err != nil {
		return zero, fmt.Errorf("object/store/packed/internal/ingest: %w", err)
	}

	return oid, nil
}

// resolvedRoots returns the indices of every currently resolved record.
func (ingestion *ingestion) resolvedRoots() []int {
	var roots []int

	for index := range ingestion.records {
		if ingestion.records[index].resolved {
			roots = append(roots, index)
		}
	}

	return roots
}

// countDeltas returns the number of delta records.
func (ingestion *ingestion) countDeltas() int {
	return ingestion.deltaCount
}

// countUnresolved returns the number of records that remain unresolved.
//
// Every base is resolved during scanning or thin completion,
// so the unresolved records are exactly the unresolved deltas:
// the delta records minus those already resolved.
func (ingestion *ingestion) countUnresolved() int {
	return ingestion.deltaCount - ingestion.deltasResolved
}

// unresolvedExternalBases returns the unique base object IDs
// of unresolved ref-deltas whose base is not present in the pack,
// in first-reference order.
func (ingestion *ingestion) unresolvedExternalBases() []id.ObjectID {
	seen := make(map[id.ObjectID]struct{})

	out := make([]id.ObjectID, 0, ingestion.deltaCount-ingestion.deltasResolved)

	for index := range ingestion.records {
		rec := &ingestion.records[index]
		if rec.resolved || rec.packedType != packfile.EntryTypeRefDelta {
			continue
		}

		if _, ok := ingestion.byOID[rec.baseOID]; ok {
			continue
		}

		if _, ok := seen[rec.baseOID]; ok {
			continue
		}

		seen[rec.baseOID] = struct{}{}
		out = append(out, rec.baseOID)
	}

	return out
}
