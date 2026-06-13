package packed

import (
	"bytes"
	"errors"
	"fmt"
	"io"

	"lindenii.org/go/furgit/internal/compress/zlib"
	"lindenii.org/go/furgit/internal/format/packfile"
	"lindenii.org/go/furgit/object/id"
	"lindenii.org/go/lgo/intconv"
)

var errPayloadOverlong = errors.New("entry payload longer than declared")

// entryHeaderAt parses the entry header at offset,
// returning it together with the entry's compressed payload view.
//
// The entry header only contains the inflated length,
// so payload slice extends to the end of the pack;
// the compressed data length is determined by the zlib stream end,
// not the slice length.
//
// Labels: Life-Parent, Mut-No.
func (pack *pack) entryHeaderAt(offset int, objectFormat id.ObjectFormat) (packfile.EntryHeader, []byte, error) {
	var zero packfile.EntryHeader

	pos := offset
	if pos < 0 || pos >= len(pack.data) {
		return zero, nil, fmt.Errorf("%w: pack %q: entry offset out of bounds", ErrMalformedPackedStore, pack.name)
	}

	header, err := packfile.ParseEntryHeader(pack.data[pos:], objectFormat.Size())
	if err != nil {
		return zero, nil, fmt.Errorf("%w: pack %q: %w", ErrMalformedPackedStore, pack.name, err)
	}

	return header, pack.data[pos+header.HeaderLen:], nil
}

// inflate decompresses one entry payload of expectedSize bytes,
// rejecting payloads whose inflated size differs.
//
// Labels: Life-Independent.
func inflate(payload []byte, expectedSize uint64) ([]byte, error) {
	size, err := intconv.Uint64ToInt(expectedSize)
	if err != nil {
		return nil, fmt.Errorf("declared size: %w", err)
	}

	zr, err := zlib.NewReader(bytes.NewReader(payload))
	if err != nil {
		return nil, fmt.Errorf("inflating entry payload: %w", err)
	}

	defer func() { _ = zr.Close() }()

	out := make([]byte, size)

	_, err = io.ReadFull(zr, out)
	if err != nil {
		return nil, fmt.Errorf("inflating entry payload: %w", err)
	}

	var probe [1]byte

	n, err := zr.Read(probe[:])
	if n != 0 || !errors.Is(err, io.EOF) {
		return nil, errPayloadOverlong
	}

	return out, nil
}
