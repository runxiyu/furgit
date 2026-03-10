package ingest

import (
	"bufio"
	"bytes"
	"errors"
	"io"
	"os"

	"codeberg.org/lindenii/furgit/objectid"
	"codeberg.org/lindenii/furgit/objectstore"
)

// Options controls one pack ingest operation.
type Options struct {
	// FixThin appends missing local bases for thin packs.
	FixThin bool
	// WriteRev writes a .rev alongside the .pack and .idx.
	WriteRev bool
	// Base supplies existing objects for thin-pack fixup.
	Base objectstore.Store
	// Progress receives human-readable progress messages.
	//
	// When nil, no progress output is emitted.
	Progress io.Writer
	// ProgressFlush flushes transport output after progress writes.
	//
	// When nil, no explicit flush is attempted.
	ProgressFlush func() error
	// RequireTrailingEOF requires the source to hit EOF after the pack trailer.
	//
	// This is suitable for exact pack-file readers, but should be disabled for
	// full-duplex transport streams like receive-pack where the peer keeps the
	// connection open to read the server response.
	RequireTrailingEOF bool
}

// Result describes one successful ingest transaction.
type Result struct {
	// PackName is the destination-relative filename of the written .pack.
	PackName string
	// IdxName is the destination-relative filename of the written .idx.
	IdxName string
	// RevName is the destination-relative filename of the written .rev.
	//
	// RevName is empty when writeRev is false.
	RevName string
	// PackHash is the final pack hash (same hash embedded in .idx/.rev trailers).
	PackHash objectid.ObjectID
	// ObjectCount is the final object count in the resulting pack.
	//
	// If thin fixup appends objects, this includes appended base objects.
	ObjectCount uint32
	// ThinFixed reports whether thin fixup appended local bases.
	ThinFixed bool
}

// HeaderInfo describes the parsed PACK header.
type HeaderInfo struct {
	Version     uint32
	ObjectCount uint32
}

// DiscardResult describes one successful Discard call.
type DiscardResult struct {
	PackHash    objectid.ObjectID
	ObjectCount uint32
}

// Pending is one started ingest operation awaiting Continue or Discard.
type Pending struct {
	reader    *bufio.Reader
	algo      objectid.Algorithm
	opts      Options
	header    HeaderInfo
	headerRaw [packHeaderSize]byte

	finalized bool
}

// Ingest reads and validates one PACK header, returning one pending operation.
func Ingest(
	src io.Reader,
	algo objectid.Algorithm,
	opts Options,
) (*Pending, error) {
	if algo.Size() == 0 {
		return nil, objectid.ErrInvalidAlgorithm
	}

	reader := bufio.NewReader(src)

	header, headerRaw, err := readAndValidatePackHeader(reader)
	if err != nil {
		return nil, err
	}

	return &Pending{
		reader:    reader,
		algo:      algo,
		opts:      opts,
		header:    header,
		headerRaw: headerRaw,
	}, nil
}

// Header returns parsed PACK header info.
func (pending *Pending) Header() HeaderInfo {
	return pending.header
}

// Continue ingests the pack stream into destination and writes pack artifacts.
func (pending *Pending) Continue(destination *os.Root) (Result, error) {
	if pending.finalized {
		return Result{}, ErrAlreadyFinalized
	}

	pending.finalized = true

	if pending.header.ObjectCount == 0 {
		return Result{}, ErrZeroObjectContinue
	}

	state, err := newIngestState(
		pending.reader,
		destination,
		pending.algo,
		pending.opts,
		pending.header,
		pending.headerRaw,
	)
	if err != nil {
		return Result{}, err
	}

	return ingest(state)
}

// Discard consumes and verifies one zero-object pack stream without writing files.
func (pending *Pending) Discard() (DiscardResult, error) {
	if pending.finalized {
		return DiscardResult{}, ErrAlreadyFinalized
	}

	pending.finalized = true

	if pending.header.ObjectCount != 0 {
		return DiscardResult{}, ErrNonZeroDiscard
	}

	hashImpl, err := pending.algo.New()
	if err != nil {
		return DiscardResult{}, err
	}

	_, _ = hashImpl.Write(pending.headerRaw[:])

	trailer := make([]byte, pending.algo.Size())

	_, err = io.ReadFull(pending.reader, trailer)
	if err != nil {
		return DiscardResult{}, &PackTrailerMismatchError{}
	}

	computed := hashImpl.Sum(nil)
	if !bytes.Equal(computed, trailer) {
		return DiscardResult{}, &PackTrailerMismatchError{}
	}

	if pending.opts.RequireTrailingEOF {
		var probe [1]byte

		n, err := pending.reader.Read(probe[:])
		if n > 0 || err == nil {
			return DiscardResult{}, errors.New("packfile/ingest: pack has trailing garbage")
		}

		if err != io.EOF {
			return DiscardResult{}, err
		}
	}

	packHash, err := objectid.FromBytes(pending.algo, trailer)
	if err != nil {
		return DiscardResult{}, err
	}

	return DiscardResult{
		PackHash:    packHash,
		ObjectCount: 0,
	}, nil
}
