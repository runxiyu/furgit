package furgit

import (
	"crypto/rand"
	"crypto/sha1"
	"crypto/sha256"
	"encoding/binary"
	"errors"
	"hash"
	"io"

	"codeberg.org/lindenii/furgit/internal/zlib"
)

// TODO
var errPackDeltaUnimplemented = errors.New("furgit: pack: delta writing not implemented")

// packWriter writes a PACKv2 stream.
type packWriter struct {
	w            io.Writer
	h            hash.Hash
	algo         hashAlgorithm
	objCount     uint32
	wroteHeader  bool
	bytesWritten uint64
}

func newPackWriter(w io.Writer, algo hashAlgorithm, objCount uint32) (*packWriter, error) {
	if w == nil {
		return nil, ErrInvalidObject
	}
	h, err := newHashWriter(algo)
	if err != nil {
		return nil, err
	}
	return &packWriter{
		w:        w,
		h:        h,
		algo:     algo,
		objCount: objCount,
	}, nil
}

func newHashWriter(algo hashAlgorithm) (hash.Hash, error) {
	switch algo {
	case hashAlgoSHA1:
		return sha1.New(), nil
	case hashAlgoSHA256:
		return sha256.New(), nil
	default:
		return nil, ErrInvalidObject
	}
}

func (pw *packWriter) writePacked(p []byte) error {
	if len(p) == 0 {
		return nil
	}
	n, err := pw.w.Write(p)
	if n > 0 {
		_, _ = pw.h.Write(p[:n])
		pw.bytesWritten += uint64(n)
	}
	if err != nil {
		return err
	}
	if n != len(p) {
		return io.ErrShortWrite
	}
	return nil
}

func (pw *packWriter) WriteHeader() error {
	if pw == nil || pw.wroteHeader {
		return ErrInvalidObject
	}
	var hdr [12]byte
	binary.BigEndian.PutUint32(hdr[0:4], packMagic)
	binary.BigEndian.PutUint32(hdr[4:8], packVersion2)
	binary.BigEndian.PutUint32(hdr[8:12], pw.objCount)
	if err := pw.writePacked(hdr[:]); err != nil {
		return err
	}
	pw.wroteHeader = true
	return nil
}

func (pw *packWriter) WriteObject(ty ObjectType, body []byte) error {
	if pw == nil || !pw.wroteHeader {
		return ErrInvalidObject
	}
	switch ty {
	case ObjectTypeCommit, ObjectTypeTree, ObjectTypeBlob, ObjectTypeTag:
		// remember that go switches don't fallthrough lol
	default:
		return ErrInvalidObject
	}
	if body == nil {
		body = []byte{}
	}

	hdr, err := packHeaderEncode(ty, len(body))
	if err != nil {
		return err
	}
	if err := pw.writePacked(hdr); err != nil {
		return err
	}

	zw := zlib.NewWriter(&packHashWriter{pw: pw})
	if _, err := zw.Write(body); err != nil {
		_ = zw.Close()
		return err
	}
	return zw.Close()
}

func (pw *packWriter) WriteOfsDelta(baseOffset uint64, baseSize, resultSize int, delta []byte) error {
	if pw == nil || !pw.wroteHeader {
		return ErrInvalidObject
	}
	if baseSize < 0 || resultSize < 0 {
		return ErrInvalidObject
	}
	if delta == nil {
		delta = []byte{}
	}
	deltaSize := len(delta)
	if deltaSize <= 0 {
		return ErrInvalidObject
	}
	currentOffset := pw.bytesWritten
	if baseOffset >= currentOffset {
		return ErrInvalidObject
	}
	dist := currentOffset - baseOffset

	hdr, err := packHeaderEncode(ObjectTypeOfsDelta, deltaSize)
	if err != nil {
		return err
	}
	if err := pw.writePacked(hdr); err != nil {
		return err
	}
	ofs, err := packOfsEncode(dist)
	if err != nil {
		return err
	}
	if err := pw.writePacked(ofs); err != nil {
		return err
	}

	zw := zlib.NewWriter(&packHashWriter{pw: pw})
	if _, err := zw.Write(delta); err != nil {
		_ = zw.Close()
		return err
	}
	return zw.Close()
}

func (pw *packWriter) WriteRefDelta(base Hash, baseSize, resultSize int, delta []byte) error {
	if pw == nil || !pw.wroteHeader {
		return ErrInvalidObject
	}
	if baseSize < 0 || resultSize < 0 {
		return ErrInvalidObject
	}
	if delta == nil {
		delta = []byte{}
	}
	deltaSize := len(delta)
	if deltaSize <= 0 {
		return ErrInvalidObject
	}
	baseBytes := base.Bytes()
	if len(baseBytes) == 0 {
		return ErrInvalidObject
	}

	hdr, err := packHeaderEncode(ObjectTypeRefDelta, deltaSize)
	if err != nil {
		return err
	}
	if err := pw.writePacked(hdr); err != nil {
		return err
	}
	if err := pw.writePacked(baseBytes); err != nil {
		return err
	}

	zw := zlib.NewWriter(&packHashWriter{pw: pw})
	if _, err := zw.Write(delta); err != nil {
		_ = zw.Close()
		return err
	}
	return zw.Close()
}

func (pw *packWriter) Close() (Hash, error) {
	if pw == nil || !pw.wroteHeader {
		return Hash{}, ErrInvalidObject
	}
	sum := pw.h.Sum(nil)
	if _, err := pw.w.Write(sum); err != nil {
		return Hash{}, err
	}
	var out Hash
	copy(out.data[:], sum)
	out.algo = pw.algo
	return out, nil
}

type packHashWriter struct {
	pw *packWriter
}

func (w *packHashWriter) Write(p []byte) (int, error) {
	if w == nil || w.pw == nil {
		return 0, ErrInvalidObject
	}
	if err := w.pw.writePacked(p); err != nil {
		return 0, err
	}
	return len(p), nil
}

// packHeaderEncode encodes a pack object header (type + size).
func packHeaderEncode(ty ObjectType, size int) ([]byte, error) {
	if size < 0 {
		return nil, ErrInvalidObject
	}
	var out [16]byte
	pos := 0

	b := byte(size & 0x0f)
	size >>= 4
	b |= byte(ty&0x07) << 4
	if size > 0 {
		b |= 0x80
	}
	out[pos] = b
	pos++

	for size > 0 {
		b = byte(size & 0x7f)
		size >>= 7
		if size > 0 {
			b |= 0x80
		}
		out[pos] = b
		pos++
	}

	return out[:pos], nil
}

// packVarintEncode encodes a 7-bit varint.
func packVarintEncode(size int) ([]byte, error) {
	if size < 0 {
		return nil, ErrInvalidObject
	}
	var out [16]byte
	pos := 0
	for {
		b := byte(size & 0x7f)
		size >>= 7
		if size != 0 {
			b |= 0x80
		}
		out[pos] = b
		pos++
		if size == 0 {
			break
		}
	}
	return out[:pos], nil
}

// packOfsEncode encodes an ofs-delta distance.
func packOfsEncode(dist uint64) ([]byte, error) {
	if dist == 0 {
		return nil, ErrInvalidObject
	}
	var out [16]byte
	pos := 0
	out[pos] = byte(dist & 0x7f)
	pos++
	dist >>= 7
	for dist != 0 {
		b := byte((dist - 1) & 0x7f)
		out[pos] = b | 0x80
		pos++
		dist >>= 7
	}
	for i, j := 0, pos-1; i < j; i, j = i+1, j-1 {
		out[i], out[j] = out[j], out[i]
	}
	return out[:pos], nil
}

// packWrite writes a pack stream for the provided object ids.
func (repo *Repository) packWrite(w io.Writer, objects []Hash, opts packWriteOptions) (Hash, error) {
	if repo == nil {
		return Hash{}, ErrInvalidObject
	}
	if opts.EnableThinPack {
		return Hash{}, errPackDeltaUnimplemented
	}
	if len(objects) > int(^uint32(0)) {
		return Hash{}, ErrInvalidObject
	}

	pw, err := newPackWriter(w, repo.hashAlgo, uint32(len(objects)))
	if err != nil {
		return Hash{}, err
	}
	if err := pw.WriteHeader(); err != nil {
		return Hash{}, err
	}

	var dctx deltaContext
	var deltaSeed uint32
	if opts.EnableDeltas {
		dctx.window = defaultDeltaWindow
		var seedBytes [4]byte
		if _, err := rand.Read(seedBytes[:]); err != nil {
			return Hash{}, err
		}
		deltaSeed = binary.LittleEndian.Uint32(seedBytes[:])
	}

	for _, id := range objects {
		ty, body, err := repo.ReadObjectTypeRaw(id)
		if err != nil {
			return Hash{}, err
		}
		obj := &objectToPack{
			id:   id,
			ty:   ty,
			body: body,
		}
		startOffset := pw.bytesWritten
		wroteDelta := false

		if opts.EnableDeltas && ty == ObjectTypeBlob {
			base, delta := pickDeltaBase(&dctx, obj, deltaSeed, opts.MinDeltaSavings, opts.MaxDeltaDepth)
			if base != nil && delta != nil {
				if err := pw.WriteOfsDelta(base.offset, len(base.body), len(body), delta); err != nil {
					return Hash{}, err
				}
				wroteDelta = true
				obj.deltaDepth = base.deltaDepth + 1
			}
		}
		if !wroteDelta {
			if err := pw.WriteObject(ty, body); err != nil {
				return Hash{}, err
			}
			obj.deltaDepth = 0
		}
		obj.offset = startOffset

		if opts.EnableDeltas && ty == ObjectTypeBlob {
			dctx.addCandidate(obj)
		}
	}

	return pw.Close()
}

type packWriteOptions struct {
	EnableDeltas    bool
	EnableThinPack  bool
	MinDeltaSavings int
	MaxDeltaDepth   int
}
