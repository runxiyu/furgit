package objectstore

// Writer represents a store that could perform both pack ingestions
// and individual object writes.
type Writer interface {
	PackWriter
	ObjectWriter
}
