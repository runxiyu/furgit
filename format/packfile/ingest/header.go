package ingest

import (
	"encoding/binary"
	"fmt"
	"io"

	"codeberg.org/lindenii/furgit/format/packfile"
)

const packHeaderSize = 12

// readAndValidatePackHeader reads one PACK header from src and validates it.
func readAndValidatePackHeader(src io.Reader) (HeaderInfo, [packHeaderSize]byte, error) {
	var hdr [packHeaderSize]byte

	_, err := io.ReadFull(src, hdr[:])
	if err != nil {
		return HeaderInfo{}, [packHeaderSize]byte{}, &InvalidPackHeaderError{
			Reason: fmt.Sprintf("read header: %v", err),
		}
	}

	header, err := parseAndValidatePackHeader(hdr)
	if err != nil {
		return HeaderInfo{}, [packHeaderSize]byte{}, err
	}

	return header, hdr, nil
}

// parseAndValidatePackHeader validates one already-read PACK header.
func parseAndValidatePackHeader(hdr [packHeaderSize]byte) (HeaderInfo, error) {
	if binary.BigEndian.Uint32(hdr[:4]) != packfile.Signature {
		return HeaderInfo{}, &InvalidPackHeaderError{Reason: "signature mismatch"}
	}

	version := binary.BigEndian.Uint32(hdr[4:8])
	if !packfile.SupportedVersion(version) {
		return HeaderInfo{}, &InvalidPackHeaderError{
			Reason: fmt.Sprintf("unsupported version %d", version),
		}
	}

	return HeaderInfo{
		Version:     version,
		ObjectCount: binary.BigEndian.Uint32(hdr[8:12]),
	}, nil
}
