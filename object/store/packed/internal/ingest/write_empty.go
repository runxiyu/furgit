package ingest

import (
	"bytes"
	"errors"
	"io"

	objectid "lindenii.org/go/furgit/object/id"
)

func discardZeroObjectPack(
	src io.Reader,
	algo objectid.Algorithm,
	opts Options,
	headerRaw [packHeaderSize]byte,
) (Result, error) {
	hashImpl, err := algo.New()
	if err != nil {
		return Result{}, err
	}

	_, _ = hashImpl.Write(headerRaw[:])

	trailer := make([]byte, algo.Size())

	_, err = io.ReadFull(src, trailer)
	if err != nil {
		return Result{}, &PackTrailerMismatchError{}
	}

	computed := hashImpl.Sum(nil)
	if !bytes.Equal(computed, trailer) {
		return Result{}, &PackTrailerMismatchError{}
	}

	if opts.RequireTrailingEOF {
		var probe [1]byte

		n, err := src.Read(probe[:])
		if n > 0 || err == nil {
			return Result{}, errors.New("packfile/ingest: pack has trailing garbage")
		}

		if err != io.EOF {
			return Result{}, err
		}
	}

	packHash, err := objectid.FromBytes(algo, trailer)
	if err != nil {
		return Result{}, err
	}

	return Result{
		PackHash:    packHash,
		ObjectCount: 0,
	}, nil
}
