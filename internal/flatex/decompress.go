package flatex

import (
	"io"

	"git.sr.ht/~runxiyu/furgit/internal/bufpool"
)

func DecompressSized(src []byte, sizeHint int) (bufpool.Buffer, int, error) {
	d := sliceInflaterPool.Get().(*sliceInflater)
	defer sliceInflaterPool.Put(d)

	if err := d.reset(src); err != nil {
		return bufpool.Buffer{}, 0, err
	}

	out := bufpool.Borrow(sizeHint)
	out.Resize(0)

	for {
		if len(d.toRead) > 0 {
			out.Append(d.toRead)
			d.toRead = nil
			continue
		}
		if d.err != nil {
			if d.err == io.EOF {
				return out, d.pos, nil
			}
			out.Release()
			return bufpool.Buffer{}, 0, d.err
		}
		d.step(d)
		if d.err != nil && len(d.toRead) == 0 {
			d.toRead = d.window.readFlush()
		}
	}
}
