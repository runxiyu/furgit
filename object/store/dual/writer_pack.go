package dual

import (
	"io"

	objectstore "codeberg.org/lindenii/furgit/object/store"
)

// WritePack ingests one pack stream into the pack-wise store.
func (dual *Dual) WritePack(src io.Reader, opts objectstore.PackWriteOptions) error {
	return dual.pack.WritePack(src, opts)
}
