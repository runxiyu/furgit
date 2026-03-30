package ingest

import (
	"codeberg.org/lindenii/furgit/common/iowrap"
	objectstore "codeberg.org/lindenii/furgit/object/store"
)

// Options controls one pack ingest operation.
type Options struct {
	// FixThin appends missing local bases for thin packs.
	FixThin bool
	// WriteRev writes a .rev alongside the .pack and .idx.
	WriteRev bool
	// Base supplies existing objects for thin-pack fixup.
	Base objectstore.Reader
	// Progress receives human-readable progress messages.
	//
	// When nil, no progress output is emitted.
	Progress iowrap.WriteFlusher
	// RequireTrailingEOF requires the source to hit EOF after the pack trailer.
	//
	// This is suitable for exact pack-file readers, but should be disabled for
	// full-duplex transport streams like receive-pack where the peer keeps the
	// connection open to read the server response.
	RequireTrailingEOF bool
}
