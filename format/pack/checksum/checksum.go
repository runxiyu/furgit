// Package checksum provides Git pack/index checksum primitives.
package checksum

import (
	"bytes"
	"fmt"

	"codeberg.org/lindenii/furgit/objectid"
)

// VerifyPackTrailer verifies one pack trailer hash against the pack payload.
//
// This computes the object hash over all bytes except the trailing hash, so it
// is O(pack size).
func VerifyPackTrailer(data []byte, algo objectid.Algorithm) error {
	hashSize := algo.Size()
	if hashSize <= 0 {
		return objectid.ErrInvalidAlgorithm
	}
	if len(data) < hashSize {
		return fmt.Errorf("format/pack/checksum: pack too short for trailer hash")
	}

	hash, err := algo.New()
	if err != nil {
		return err
	}
	if _, err := hash.Write(data[:len(data)-hashSize]); err != nil {
		return err
	}
	computed := hash.Sum(nil)
	trailer := data[len(data)-hashSize:]
	if !bytes.Equal(computed, trailer) {
		return fmt.Errorf("format/pack/checksum: pack trailer hash mismatch")
	}
	return nil
}

// PackTrailerHash returns the trailer hash bytes from one pack file.
func PackTrailerHash(data []byte, algo objectid.Algorithm) ([]byte, error) {
	hashSize := algo.Size()
	if hashSize <= 0 {
		return nil, objectid.ErrInvalidAlgorithm
	}
	if len(data) < hashSize {
		return nil, fmt.Errorf("format/pack/checksum: pack too short for trailer hash")
	}
	return data[len(data)-hashSize:], nil
}

// ParseIdxTrailer parses one idx v2 trailer and returns (packHash, idxHash).
func ParseIdxTrailer(data []byte, algo objectid.Algorithm) ([]byte, []byte, error) {
	hashSize := algo.Size()
	if hashSize <= 0 {
		return nil, nil, objectid.ErrInvalidAlgorithm
	}
	if len(data) < hashSize*2 {
		return nil, nil, fmt.Errorf("format/pack/checksum: idx too short for trailer hashes")
	}
	packHashOff := len(data) - hashSize*2
	idxHashOff := len(data) - hashSize
	return data[packHashOff:idxHashOff], data[idxHashOff:], nil
}

// VerifyIdxTrailer verifies one idx trailer checksum against preceding bytes.
func VerifyIdxTrailer(data []byte, algo objectid.Algorithm) error {
	hashSize := algo.Size()
	if hashSize <= 0 {
		return objectid.ErrInvalidAlgorithm
	}
	if len(data) < hashSize*2 {
		return fmt.Errorf("format/pack/checksum: idx too short for trailer hashes")
	}

	_, idxHash, err := ParseIdxTrailer(data, algo)
	if err != nil {
		return err
	}
	hash, err := algo.New()
	if err != nil {
		return err
	}
	if _, err := hash.Write(data[:len(data)-hashSize]); err != nil {
		return err
	}
	computed := hash.Sum(nil)
	if !bytes.Equal(computed, idxHash) {
		return fmt.Errorf("format/pack/checksum: idx trailer hash mismatch")
	}
	return nil
}

// VerifyPackMatchesIdx compares a pack trailer hash with one idx-recorded pack
// hash.
func VerifyPackMatchesIdx(packData, idxData []byte, algo objectid.Algorithm) error {
	packTrailerHash, err := PackTrailerHash(packData, algo)
	if err != nil {
		return err
	}
	idxPackHash, _, err := ParseIdxTrailer(idxData, algo)
	if err != nil {
		return err
	}
	if !bytes.Equal(packTrailerHash, idxPackHash) {
		return fmt.Errorf("format/pack/checksum: pack hash does not match idx")
	}
	return nil
}
