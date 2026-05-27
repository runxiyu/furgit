package ingest

import (
	"bufio"
	"io"
	"os"

	objectid "lindenii.org/go/furgit/object/id"
)

// WritePack ingests one pack stream into destination and writes pack artifacts.
//
// Artifacts are published under content-addressed final names derived from the
// resulting pack hash. If those final names already exist, WritePack treats
// that as success and removes its temporary files.
func WritePack(
	destination *os.Root,
	algo objectid.Algorithm,
	src io.Reader,
	opts Options,
) (Result, error) {
	if algo.Size() == 0 {
		return Result{}, objectid.ErrInvalidAlgorithm
	}

	reader := bufio.NewReader(src)

	header, headerRaw, err := readAndValidatePackHeader(reader)
	if err != nil {
		return Result{}, err
	}

	if header.ObjectCount == 0 {
		return discardZeroObjectPack(reader, algo, opts, headerRaw)
	}

	state, err := newIngestState(
		reader,
		destination,
		algo,
		opts,
		header,
		headerRaw,
	)
	if err != nil {
		return Result{}, err
	}

	return ingest(state)
}
