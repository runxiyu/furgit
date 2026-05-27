package signedcommit

import objectid "lindenii.org/go/furgit/object/id"

// Commit represents the payload and signatures parsed from a raw comit object.
type Commit struct {
	body       []byte
	payload    []byteRange
	signatures map[objectid.Algorithm][]byteRange
}

type byteRange struct {
	start int
	end   int
}
