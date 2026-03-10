package ingest

import (
	"encoding/binary"
	"io"

	"codeberg.org/lindenii/furgit/internal/intconv"
	"codeberg.org/lindenii/furgit/objectid"
)

// rewritePackHeaderAndTrailer rewrites object count and trailer hash using ReadAt/WriteAt.
func rewritePackHeaderAndTrailer(state *ingestState) error {
	var countRaw [4]byte

	recordCountUint32, err := intconv.IntToUint32(len(state.records))
	if err != nil {
		return err
	}

	binary.BigEndian.PutUint32(countRaw[:], recordCountUint32)

	_, err = state.packFile.WriteAt(countRaw[:], 8)
	if err != nil {
		return err
	}

	info, err := state.packFile.Stat()
	if err != nil {
		return err
	}

	endWithoutTrailer := info.Size()

	hashImpl, err := state.algo.New()
	if err != nil {
		return err
	}

	var (
		buf [128 << 10]byte
		pos int64
	)
	for pos < endWithoutTrailer {
		want := int64(len(buf))

		remaining := endWithoutTrailer - pos
		if remaining < want {
			want = remaining
		}

		n, err := state.packFile.ReadAt(buf[:want], pos)
		if err != nil && err != io.EOF {
			return err
		}

		if n == 0 {
			return io.ErrUnexpectedEOF
		}

		_, _ = hashImpl.Write(buf[:n])
		pos += int64(n)
	}

	sum := hashImpl.Sum(nil)

	_, err = state.packFile.WriteAt(sum, endWithoutTrailer)
	if err != nil {
		return err
	}

	packHash, err := objectid.FromBytes(state.algo, sum)
	if err != nil {
		return err
	}

	state.packHash = packHash
	state.objectCountHeader = recordCountUint32

	sumLenInt64 := int64(len(sum))

	newConsumed, err := intconv.Int64ToUint64(endWithoutTrailer + sumLenInt64)
	if err != nil {
		return err
	}

	state.stream.consumed = newConsumed

	return nil
}
