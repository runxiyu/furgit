package ingest

import (
	"encoding/binary"
	"fmt"

	"codeberg.org/lindenii/furgit/format/pack"
)

// readAndValidatePackHeader reads and validates PACK header from the stream.
func readAndValidatePackHeader(state *ingestState) error {
	var hdr [12]byte

	err := state.stream.readFull(hdr[:])
	if err != nil {
		return &ErrInvalidPackHeader{Reason: fmt.Sprintf("read header: %v", err)}
	}

	if binary.BigEndian.Uint32(hdr[:4]) != pack.Signature {
		return &ErrInvalidPackHeader{Reason: "signature mismatch"}
	}

	version := binary.BigEndian.Uint32(hdr[4:8])
	if !pack.VersionSupported(version) {
		return &ErrInvalidPackHeader{Reason: fmt.Sprintf("unsupported version %d", version)}
	}

	state.objectCountHeader = binary.BigEndian.Uint32(hdr[8:12])
	if state.objectCountHeader == 0 {
		return &ErrInvalidPackHeader{Reason: "zero objects"}
	}

	return nil
}
