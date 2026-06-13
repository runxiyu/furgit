package ingest

import (
	"errors"
	"fmt"
	"io"
	"io/fs"
	"slices"

	"lindenii.org/go/furgit/internal/format/packidx"
	"lindenii.org/go/furgit/internal/format/packrev"
	"lindenii.org/go/furgit/object/id"
	"lindenii.org/go/lgo/intconv"
)

// finalize writes the index and reverse index,
// then links the pack, reverse index, and index
// to their content-addressed names.
func (ingestion *ingestion) finalize() (Result, error) {
	entries, positions, err := ingestion.indexEntries()
	if err != nil {
		return Result{}, err
	}

	packHash := ingestion.packHash.Bytes()

	idxTmp, err := ingestion.writeTemp("tmp_idx_", func(w io.Writer) error {
		return packidx.Write(w, ingestion.objectFormat, entries, packHash)
	})
	if err != nil {
		return Result{}, err
	}

	revTmp, err := ingestion.writeTemp("tmp_rev_", func(w io.Writer) error {
		return packrev.Write(w, ingestion.objectFormat, positions, packHash)
	})
	if err != nil {
		return Result{}, err
	}

	base := "pack-" + ingestion.packHash.String()
	packFinal := base + ".pack"
	idxFinal := base + ".idx"
	revFinal := base + ".rev"

	// Link the pack and reverse index before the index,
	// since the index is what publishes the pack to readers.
	err = ingestion.link(ingestion.packTmp, packFinal)
	if err != nil {
		return Result{}, err
	}

	err = ingestion.link(revTmp, revFinal)
	if err != nil {
		return Result{}, err
	}

	err = ingestion.link(idxTmp, idxFinal)
	if err != nil {
		return Result{}, err
	}

	objectCount, err := intconv.IntToUint32(len(ingestion.records))
	if err != nil {
		return Result{}, fmt.Errorf("object/store/packed/internal/ingest: %w", err)
	}

	return Result{
		PackName:    packFinal,
		IdxName:     idxFinal,
		RevName:     revFinal,
		PackHash:    ingestion.packHash,
		ObjectCount: objectCount,
		ThinFixed:   ingestion.thinFixed,
	}, nil
}

// indexEntries returns the index entries in object-ID order
// and, for each record in pack order, its position in that index order.
func (ingestion *ingestion) indexEntries() ([]packidx.Entry, []uint32, error) {
	order := make([]int, len(ingestion.records))
	for i := range order {
		order[i] = i
	}

	slices.SortFunc(order, func(left, right int) int {
		return ingestion.records[left].oid.Compare(ingestion.records[right].oid)
	})

	entries := make([]packidx.Entry, len(order))
	positions := make([]uint32, len(ingestion.records))

	for indexPosition, recordIndex := range order {
		rec := &ingestion.records[recordIndex]

		var oidBytes [id.MaxObjectIDSize]byte
		copy(oidBytes[:], rec.oid.RawBytes())

		offset, err := intconv.IntToUint64(rec.offset)
		if err != nil {
			return nil, nil, fmt.Errorf("object/store/packed/internal/ingest: %w", err)
		}

		entries[indexPosition] = packidx.Entry{
			OID:    oidBytes,
			Offset: offset,
			CRC32:  rec.crc32,
		}

		position, err := intconv.IntToUint32(indexPosition)
		if err != nil {
			return nil, nil, fmt.Errorf("object/store/packed/internal/ingest: %w", err)
		}

		positions[recordIndex] = position
	}

	return entries, positions, nil
}

// writeTemp creates a temporary file,
// writes it via write, syncs it, and returns its name.
func (ingestion *ingestion) writeTemp(prefix string, write func(io.Writer) error) (string, error) {
	name, file, err := ingestion.createTemp(prefix)
	if err != nil {
		return "", err
	}

	defer func() { _ = file.Close() }()

	err = write(file)
	if err != nil {
		return "", err
	}

	err = file.Sync()
	if err != nil {
		return "", fmt.Errorf("object/store/packed/internal/ingest: syncing %q: %w", name, err)
	}

	return name, nil
}

// link hard-links tmp to final,
// treating an already-present destination as success.
func (ingestion *ingestion) link(tmp, final string) error {
	err := ingestion.root.Link(tmp, final)
	if err != nil && !errors.Is(err, fs.ErrExist) {
		return fmt.Errorf("object/store/packed/internal/ingest: linking %q: %w", final, err)
	}

	_ = ingestion.root.Remove(tmp)

	return nil
}
