package flate

import (
	"io"
	"sync"

	"git.sr.ht/~runxiyu/furgit/internal/bufpool"
)

// byteSliceReader implements Reader over an in-memory byte slice.
type byteSliceReader struct {
	data []byte
	off  int
}

func (r *byteSliceReader) Reset(data []byte) {
	r.data = data
	r.off = 0
}

func (r *byteSliceReader) Read(p []byte) (int, error) {
	if r.off >= len(r.data) {
		return 0, io.EOF
	}
	n := copy(p, r.data[r.off:])
	r.off += n
	return n, nil
}

func (r *byteSliceReader) ReadByte() (byte, error) {
	if r.off >= len(r.data) {
		return 0, io.EOF
	}
	b := r.data[r.off]
	r.off++
	return b, nil
}

// bufferDecompressor wraps the core decompressor with pooling state so that
// byte-slice decompressions avoid repeated allocations.
type bufferDecompressor struct {
	dec    decompressor
	reader byteSliceReader
}

var bufferDecompressorPool = sync.Pool{
	New: func() any {
		fixedHuffmanDecoderInit()
		d := &bufferDecompressor{}
		d.dec.bits = new([maxNumLit + maxNumDist]int)
		d.dec.codebits = new([numCodes]int)
		d.dec.step = (*decompressor).nextBlock
		return d
	},
}

// Decompress inflates the provided DEFLATE stream and returns the full output
// in a pooled bufpool.Buffer along with the number of consumed bytes from src.
func Decompress(src []byte) (bufpool.Buffer, int, error) {
	return DecompressDict(src, nil)
}

// DecompressDict inflates the provided DEFLATE stream using dict as the preset
// dictionary and returns the full output in a pooled bufpool.Buffer. The second
// returned value reports how many bytes of src were consumed.
func DecompressDict(src []byte, dict []byte) (bufpool.Buffer, int, error) {
	d := bufferDecompressorPool.Get().(*bufferDecompressor)
	defer func() {
		d.reader.Reset(nil)
		bufferDecompressorPool.Put(d)
	}()

	d.reader.Reset(src)
	if err := d.dec.Reset(&d.reader, dict); err != nil {
		return bufpool.Buffer{}, 0, err
	}

	out := bufpool.Borrow(bufpool.DefaultBufferCap)
	out.Resize(0)

	for {
		if len(d.dec.toRead) > 0 {
			out.Append(d.dec.toRead)
			d.dec.toRead = nil
			continue
		}
		if d.dec.err != nil {
			if d.dec.err == io.EOF {
				return out, d.reader.off, nil
			}
			out.Release()
			return bufpool.Buffer{}, 0, d.dec.err
		}
		d.dec.step(&d.dec)
		if d.dec.err != nil && len(d.dec.toRead) == 0 {
			d.dec.toRead = d.dec.dict.readFlush()
		}
	}
}
