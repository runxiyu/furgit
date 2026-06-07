package commit

import "lindenii.org/go/furgit/object/id"

// Commit represents the payload and signatures
// parsed from a raw commit object.
type Commit struct {
	body       []byte
	payload    []byteRange
	signatures map[id.ObjectFormat][]byteRange
}

// ObjectFormats returns the object formats
// for which the commit carries signatures.
func (commit *Commit) ObjectFormats() []id.ObjectFormat {
	var objectFormats []id.ObjectFormat

	for _, objectFormat := range id.SupportedObjectFormats() {
		if _, ok := commit.signatures[objectFormat]; ok {
			objectFormats = append(objectFormats, objectFormat)
		}
	}

	return objectFormats
}

type byteRange struct {
	start int
	end   int
}
