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
	"lindenii.org/go/lgo/intconv"
)

// adjacency maps each resolvable base to its delta children:
// ofs-deltas keyed by base offset, ref-deltas keyed by base object ID.
type adjacency struct {
	byOffset map[uint64][]int
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
		byOffset: make(map[uint64][]int),
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

// resolveFrom resolves the delta subtree rooted at each resolved record.
func (ingestion *ingestion) resolveFrom(roots []int, adjacency adjacency, meter *progress.Meter) error {
	for _, root := range roots {
		content, err := ingestion.inflateRecord(root)
		if err != nil {
			return err
		}

		err = ingestion.resolveSubtree(root, content, ingestion.records[root].objectType, 0, adjacency, meter)
		if err != nil {
			return err
		}
	}

	return nil
}

// resolveSubtree resolves every delta child of one resolved record at depth,
// holding the record's content as the base for its children.
func (ingestion *ingestion) resolveSubtree(
	index int,
	content []byte,
	objectType packfile.EntryType,
	depth int,
	adjacency adjacency,
	meter *progress.Meter,
) error {
	rec := &ingestion.records[index]

	for _, child := range adjacency.byOffset[rec.offset] {
		err := ingestion.resolveChild(child, content, objectType, depth+1, adjacency, meter)
		if err != nil {
			return err
		}
	}

	for _, child := range adjacency.byOID[rec.oid] {
		err := ingestion.resolveChild(child, content, objectType, depth+1, adjacency, meter)
		if err != nil {
			return err
		}
	}

	return nil
}

// resolveChild applies one delta record at depth against its base content,
// finalizes the record, and recurses into its own children.
func (ingestion *ingestion) resolveChild(
	index int,
	baseContent []byte,
	baseType packfile.EntryType,
	depth int,
	adjacency adjacency,
	meter *progress.Meter,
) error {
	rec := &ingestion.records[index]
	if rec.resolved {
		return nil
	}

	if depth > delta.MaxChainDepth {
		return fmt.Errorf("%w: entry at %d: delta chain too deep", ErrMalformedPack, rec.offset)
	}

	deltaPayload, err := ingestion.inflateRecord(index)
	if err != nil {
		return err
	}

	baseSize, resultSize, _, err := delta.ParseHeaderSizes(deltaPayload)
	if err != nil {
		return fmt.Errorf("%w: entry at %d: %w", ErrMalformedPack, rec.offset, err)
	}

	if baseSize != uint64(len(baseContent)) {
		return fmt.Errorf("%w: entry at %d: delta base size mismatch", ErrMalformedPack, rec.offset)
	}

	content, err := delta.Apply(baseContent, deltaPayload)
	if err != nil {
		return fmt.Errorf("%w: entry at %d: %w", ErrMalformedPack, rec.offset, err)
	}

	if uint64(len(content)) != resultSize {
		return fmt.Errorf("%w: entry at %d: delta result size mismatch", ErrMalformedPack, rec.offset)
	}

	oid, err := ingestion.hashObject(baseType, content)
	if err != nil {
		return err
	}

	rec.objectType = baseType
	rec.oid = oid
	rec.resolved = true
	ingestion.byOID[oid] = index

	ingestion.deltasResolved++
	meter.Set(ingestion.deltasResolved, 0)

	return ingestion.resolveSubtree(index, content, baseType, depth, adjacency, meter)
}

// inflateRecord inflates one record's payload from the temporary pack file.
func (ingestion *ingestion) inflateRecord(index int) ([]byte, error) {
	rec := &ingestion.records[index]

	offset, err := intconv.Uint64ToInt64(rec.dataOffset())
	if err != nil {
		return nil, fmt.Errorf("object/store/packed/internal/ingest: %w", err)
	}

	compressedLen, err := intconv.Uint64ToInt64(rec.packedLen - rec.headerLen)
	if err != nil {
		return nil, fmt.Errorf("object/store/packed/internal/ingest: %w", err)
	}

	size, err := intconv.Uint64ToInt(rec.declaredSize)
	if err != nil {
		return nil, fmt.Errorf("object/store/packed/internal/ingest: %w", err)
	}

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
func (ingestion *ingestion) hashObject(objectType packfile.EntryType, content []byte) (id.ObjectID, error) {
	var zero id.ObjectID

	ty, err := objectType.ObjectType()
	if err != nil {
		return zero, fmt.Errorf("object/store/packed/internal/ingest: %w", err)
	}

	hashImpl, err := ingestion.objectFormat.New()
	if err != nil {
		return zero, fmt.Errorf("object/store/packed/internal/ingest: %w", err)
	}

	_, _ = hashImpl.Write(header.Append(nil, ty, uint64(len(content))))
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
func (ingestion *ingestion) countDeltas() uint64 {
	return ingestion.deltaCount
}

// countUnresolved returns the number of records that remain unresolved.
//
// Every base is resolved during scanning or thin completion,
// so the unresolved records are exactly the unresolved deltas:
// the delta records minus those already resolved.
func (ingestion *ingestion) countUnresolved() uint64 {
	return ingestion.deltaCount - ingestion.deltasResolved
}

// unresolvedExternalBases returns the unique base object IDs
// of unresolved ref-deltas whose base is not present in the pack,
// in first-reference order.
func (ingestion *ingestion) unresolvedExternalBases() []id.ObjectID {
	seen := make(map[id.ObjectID]struct{})

	var out []id.ObjectID

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
