package flatex

import (
	"io"
	"sync"

	"git.sr.ht/~runxiyu/furgit/internal/bufpool"
)

type bufferDecompressor struct {
	inflater sliceInflater
}

var bufferDecompressorPool = sync.Pool{
	New: func() any {
		fixedHuffmanDecoderInit()
		d := &bufferDecompressor{
			inflater: sliceInflater{
				bits:     new([maxNumLit + maxNumDist]int),
				codebits: new([numCodes]int),
			},
		}
		return d
	},
}

func DecompressSized(src []byte, sizeHint int) (bufpool.Buffer, int, error) {
	d := bufferDecompressorPool.Get().(*bufferDecompressor)
	defer bufferDecompressorPool.Put(d)

	if err := d.inflater.reset(src); err != nil {
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
			d.inflater.toRead = d.inflater.window.readFlush()
		}
	}
}
