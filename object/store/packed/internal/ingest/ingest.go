package ingest

import (
	"bytes"
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"runtime"
	"sync/atomic"

	"lindenii.org/go/furgit/internal/cache/clock"
	"lindenii.org/go/furgit/internal/format/packfile"
	"lindenii.org/go/furgit/object/id"
	"lindenii.org/go/furgit/object/store"
	"lindenii.org/go/lgo/sync"
)

var errTempNamesExhausted = errors.New("object/store/packed/internal/ingest: exhausted temporary file names")

// ingestion holds the state for one WritePack call.
type ingestion struct {
	ctx context.Context //nolint:containedctx

	// root is the destination objects/pack directory.
	root *os.Root

	// objectFormat is the pack's object format.
	objectFormat id.ObjectFormat

	// opts carries the thin base reader and progress sink.
	opts store.PackWriteOptions

	// src is the pack stream, positioned just past the header.
	src io.Reader

	// packFile is the temporary pack file being written,
	// and packTmp is its destination-relative name.
	packFile *os.File
	packTmp  string

	// temps lists every temporary file to remove on failure.
	temps []string

	// scanner streams src into packFile while scanning entries.
	scanner *scanner

	// records holds one entry per object, in pack-offset order.
	records []record

	// byOffset maps an entry offset to its record index,
	// and byOID maps a resolved object ID to its record index.
	byOffset map[int]int
	byOID    sync.Map[id.ObjectID, int] //exhaustruct:optional

	baseCache *clock.Clock[baseCacheKey, cachedContent]

	// workers is the delta-resolution concurrency.
	workers int

	// headerCount is the object count declared by the pack header.
	headerCount int

	// deltaCount counts delta records, accumulated during scanning.
	deltaCount int

	// deltasResolved counts resolved delta records.
	deltasResolved atomic.Int64 //exhaustruct:optional

	// packHash is the final pack trailer hash.
	packHash id.ObjectID

	// thinFixed reports whether thin completion appended local bases.
	thinFixed bool

	// committed suppresses temporary file removal once artifacts are published.
	committed bool
}

// WritePack ingests one pack stream into root,
// publishing a pack, index, and reverse index
// under content-addressed names derived from the pack trailer hash.
//
// WritePack consumes the pack stream through its trailer and stops there.
// It does not require src to reach EOF afterward,
// so it is safe on a still-open transport connection,
// such as receive-pack,
// whose peer keeps the connection open to read the response.
//
// The pack must be the last thing the peer sends before that response:
// any bytes arriving immediately after the trailer
// are rejected as a malformed pack.
func WritePack(ctx context.Context, root *os.Root, objectFormat id.ObjectFormat, src io.Reader, opts store.PackWriteOptions) (Result, error) {
	if objectFormat.Size() == 0 {
		return Result{}, id.ErrInvalidObjectFormat
	}

	headerRaw, count, err := readPackHeader(src)
	if err != nil {
		return Result{}, err
	}

	workers := runtime.GOMAXPROCS(0)

	ingestion := &ingestion{
		ctx:          ctx,
		root:         root,
		objectFormat: objectFormat,
		opts:         opts,
		src:          src,
		packFile:     nil,
		packTmp:      "",
		temps:        nil,
		scanner:      nil,
		records:      nil,
		byOffset:     make(map[int]int),
		baseCache:    newBaseCache(workers),
		workers:      workers,
		headerCount:  count,
		deltaCount:   0,
		packHash:     id.ObjectID{},
		thinFixed:    false,
		committed:    false,
	}

	defer ingestion.cleanup()

	if count == 0 {
		return ingestion.finishEmpty(headerRaw)
	}

	err = ingestion.openPackTemp(headerRaw)
	if err != nil {
		return Result{}, err
	}

	err = ingestion.streamAndScan()
	if err != nil {
		return Result{}, err
	}

	err = ingestion.resolveDeltas()
	if err != nil {
		return Result{}, err
	}

	err = ingestion.packFile.Sync()
	if err != nil {
		return Result{}, fmt.Errorf("object/store/packed/internal/ingest: syncing pack: %w", err)
	}

	result, err := ingestion.finalize()
	if err != nil {
		return Result{}, err
	}

	ingestion.committed = true

	return result, nil
}

// finishEmpty verifies the trailer of a zero-object pack
// and returns success without writing any artifacts.
func (ingestion *ingestion) finishEmpty(headerRaw [packfile.HeaderLen]byte) (Result, error) {
	hashImpl, err := ingestion.objectFormat.New()
	if err != nil {
		return Result{}, fmt.Errorf("object/store/packed/internal/ingest: %w", err)
	}

	_, _ = hashImpl.Write(headerRaw[:])

	trailer := make([]byte, ingestion.objectFormat.Size())

	_, err = io.ReadFull(ingestion.src, trailer)
	if err != nil {
		return Result{}, fmt.Errorf("%w: reading trailer: %w", ErrMalformedPack, err)
	}

	if !bytes.Equal(hashImpl.Sum(nil), trailer) {
		return Result{}, fmt.Errorf("%w: trailer hash mismatch", ErrMalformedPack)
	}

	packHash, err := ingestion.objectFormat.FromBytes(trailer)
	if err != nil {
		return Result{}, fmt.Errorf("object/store/packed/internal/ingest: %w", err)
	}

	return Result{
		PackName:    "",
		IdxName:     "",
		RevName:     "",
		BloomName:   "",
		PackHash:    packHash,
		ObjectCount: 0,
		ThinFixed:   false,
	}, nil
}

// openPackTemp creates the temporary pack file,
// writes the validated header to it,
// and builds the stream scanner seeded with that header.
func (ingestion *ingestion) openPackTemp(headerRaw [packfile.HeaderLen]byte) error {
	name, file, err := ingestion.createTemp("tmp_pack_")
	if err != nil {
		return err
	}

	ingestion.packTmp = name
	ingestion.packFile = file

	_, err = file.Write(headerRaw[:])
	if err != nil {
		return fmt.Errorf("object/store/packed/internal/ingest: writing pack header: %w", err)
	}

	hashImpl, err := ingestion.objectFormat.New()
	if err != nil {
		return fmt.Errorf("object/store/packed/internal/ingest: %w", err)
	}

	_, _ = hashImpl.Write(headerRaw[:])

	ingestion.scanner = newScanner(ingestion.src, ingestion.packFile, hashImpl)

	return nil
}

// createTemp creates one temporary file under root,
// recording its name for cleanup on failure.
func (ingestion *ingestion) createTemp(prefix string) (string, *os.File, error) {
	for range 32 {
		name := prefix + rand.Text()

		file, err := ingestion.root.OpenFile(name, os.O_CREATE|os.O_EXCL|os.O_RDWR, 0o644)
		if err == nil {
			ingestion.temps = append(ingestion.temps, name)

			return name, file, nil
		}

		if !errors.Is(err, fs.ErrExist) {
			return "", nil, fmt.Errorf("object/store/packed/internal/ingest: creating temp file: %w", err)
		}
	}

	return "", nil, errTempNamesExhausted
}

// cleanup closes the pack file
// and removes temporary files unless the write committed.
func (ingestion *ingestion) cleanup() {
	if ingestion.packFile != nil {
		_ = ingestion.packFile.Close()
	}

	if ingestion.committed {
		return
	}

	for _, name := range ingestion.temps {
		_ = ingestion.root.Remove(name)
	}
}
