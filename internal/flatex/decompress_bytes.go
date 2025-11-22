package flatex

import (
	"io"
	"sync"

	"git.sr.ht/~runxiyu/furgit/internal/bufpool"
)

// bufferDecompressor wraps the custom slice inflater so byte-slice
// decompressions avoid repeated allocations.
type bufferDecompressor struct {
	inflater sliceInflater
}

var bufferDecompressorPool = sync.Pool{
	New: func() any {
		fixedHuffmanDecoderInit()
		d := &bufferDecompressor{}
		d.inflater.bits = new([maxNumLit + maxNumDist]int)
		d.inflater.codebits = new([numCodes]int)
		return d
	},
}

// Decompress inflates the provided DEFLATE stream and returns the full output
// in a pooled bufpool.Buffer along with the number of consumed bytes from src.
func Decompress(src []byte) (bufpool.Buffer, int, error) {
	return DecompressDictSized(src, nil, 0)
}

// DecompressDict inflates the provided DEFLATE stream using dict as the preset
// dictionary and returns the full output in a pooled bufpool.Buffer. The second
// returned value reports how many bytes of src were consumed.
func DecompressDict(src []byte, dict []byte) (bufpool.Buffer, int, error) {
	return DecompressDictSized(src, dict, 0)
}

// DecompressDictSized is like DecompressDict but allows providing an expected
// output size to pre-size the destination buffer and avoid repeated growth.
// A non-positive sizeHint falls back to the default buffer capacity.
func DecompressDictSized(src []byte, dict []byte, sizeHint int) (bufpool.Buffer, int, error) {
	d := bufferDecompressorPool.Get().(*bufferDecompressor)
	defer bufferDecompressorPool.Put(d)

	if err := d.inflater.reset(src, dict); err != nil {
		return bufpool.Buffer{}, 0, err
	}

	out := bufpool.Borrow(sizeHint)
	out.Resize(0)

	for {
		if len(d.inflater.toRead) > 0 {
			out.Append(d.inflater.toRead)
			d.inflater.toRead = nil
			continue
		}
		if d.inflater.err != nil {
			if d.inflater.err == io.EOF {
				return out, d.inflater.pos, nil
			}
			out.Release()
			return bufpool.Buffer{}, 0, d.inflater.err
		}
		d.inflater.step(&d.inflater)
		if d.inflater.err != nil && len(d.inflater.toRead) == 0 {
			d.inflater.toRead = d.inflater.dict.readFlush()
		}
	}
}
