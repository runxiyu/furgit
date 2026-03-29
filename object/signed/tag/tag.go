package signedtag

import objectid "codeberg.org/lindenii/furgit/object/id"

// Tag represents the payload and signatures parsed from a raw tag object.
type Tag struct {
	body       []byte
	payload    []byteRange
	signatures map[objectid.Algorithm][]byteRange
}

type byteRange struct {
	start int
	end   int
}
