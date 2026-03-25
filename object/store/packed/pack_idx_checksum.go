package packed

import (
	"bytes"
	"fmt"

	objectid "codeberg.org/lindenii/furgit/object/id"
)

// verifyMappedPackMatchesMappedIdx compares one mapped pack trailer hash with
// the pack hash recorded in one mapped idx trailer.
func verifyMappedPackMatchesMappedIdx(packData, idxData []byte, algo objectid.Algorithm) error {
	hashSize := algo.Size()
	if hashSize <= 0 {
		return objectid.ErrInvalidAlgorithm
	}

	if len(packData) < hashSize {
		return fmt.Errorf("objectstore/packed: pack too short for trailer hash")
	}

	if len(idxData) < hashSize*2 {
		return fmt.Errorf("objectstore/packed: idx too short for trailer hashes")
	}

	packTrailerHash := packData[len(packData)-hashSize:]

	idxPackHash := idxData[len(idxData)-hashSize*2 : len(idxData)-hashSize]
	if !bytes.Equal(packTrailerHash, idxPackHash) {
		return fmt.Errorf("objectstore/packed: pack hash does not match idx")
	}

	return nil
}
