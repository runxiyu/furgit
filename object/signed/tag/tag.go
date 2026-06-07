package tag

import "lindenii.org/go/furgit/object/id"

// Tag represents the payload and signatures
// parsed from a raw tag object.
type Tag struct {
	body       []byte
	payload    []byteRange
	signatures map[id.ObjectFormat][]byteRange
}

// ObjectFormats returns the object formats
// for which the tag carries signatures.
func (tag *Tag) ObjectFormats() []id.ObjectFormat {
	var objectFormats []id.ObjectFormat

	for _, objectFormat := range id.SupportedObjectFormats() {
		if _, ok := tag.signatures[objectFormat]; ok {
			objectFormats = append(objectFormats, objectFormat)
		}
	}

	return objectFormats
}

type byteRange struct {
	start int
	end   int
}
