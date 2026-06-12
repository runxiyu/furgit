package packed

import (
	"fmt"
	"io"

	"lindenii.org/go/furgit/object/store"
	"lindenii.org/go/furgit/object/store/packed/internal/ingest"
)

var _ store.PackWriter = (*Packed)(nil)

// WritePack ingests one pack stream into the pack store,
// publishing a pack, index, and reverse index
// under content-addressed names derived from the pack trailer hash.
//
// WritePack consumes the pack stream through its trailer and stops there.
// It does not require src to reach EOF afterward,
// so it is safe on a still-open transport connection,
// such as receive-pack,
// whose peer keeps the connection open to read the response.
//
// The pack must be the last thing the peer sends before that response:
// any bytes arriving immediately after the trailer
// are rejected as a malformed pack.
func (packed *Packed) WritePack(src io.Reader, opts store.PackWriteOptions) error {
	_, err := ingest.WritePack(packed.root, packed.objectFormat, src, opts)
	if err != nil {
		return err //nolint:wrapcheck
	}

	err = packed.Refresh()
	if err != nil {
		return fmt.Errorf("object/store/packed: refresh after pack write: %w", err)
	}

	return nil
}
