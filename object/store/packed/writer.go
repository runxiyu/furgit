package packed

import (
	"io"

	objectstore "codeberg.org/lindenii/furgit/object/store"
	"codeberg.org/lindenii/furgit/object/store/packed/internal/ingest"
)

var _ objectstore.PackWriter = (*Store)(nil)

// WritePack ingests one pack stream into the packed store.
func (store *Store) WritePack(src io.Reader, _ objectstore.PackWriteOptions) error {
	_, err := ingest.WritePack(store.root, store.algo, src, ingest.Options{})

	return err
}
