package bloom

import "encoding/binary"

// Settings describe the changed-paths Bloom filter parameters stored in
// commit-graph BDAT chunks.
//
// Obviously, they must match the repository's commit-graph settings to
// interpret filters correctly.
type Settings struct {
	HashVersion    uint32
	NumHashes      uint32
	BitsPerEntry   uint32
	MaxChangePaths uint32
}

// ParseSettings reads Bloom filter settings from a BDAT chunk header.
func ParseSettings(bdat []byte) (*Settings, error) {
	if len(bdat) < DataHeaderSize {
		return nil, ErrInvalid
	}

	settings := &Settings{
		HashVersion:    binary.BigEndian.Uint32(bdat[0:4]),
		NumHashes:      binary.BigEndian.Uint32(bdat[4:8]),
		BitsPerEntry:   binary.BigEndian.Uint32(bdat[8:12]),
		MaxChangePaths: DefaultMaxChange,
	}

	return settings, nil
}
