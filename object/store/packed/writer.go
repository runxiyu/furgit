package packed

import (
	"io"

	objectstore "lindenii.org/go/furgit/object/store"
	"lindenii.org/go/furgit/object/store/packed/internal/ingest"
)

var _ objectstore.PackWriter = (*Store)(nil)

// WritePack ingests one pack stream into the packed store.
func (store *Store) WritePack(src io.Reader, opts objectstore.PackWriteOptions) error {
	_, err := ingest.WritePack(store.root, store.algo, src, ingest.Options{
		WriteRev:           store.opts.WriteRev,
		FixThin:            opts.ThinBase != nil,
		ThinBase:           opts.ThinBase,
		Progress:           opts.Progress,
		RequireTrailingEOF: opts.RequireTrailingEOF,
	})

	return err
}
