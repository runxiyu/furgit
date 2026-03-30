package reading

// location identifies one object entry in a specific pack file.
type location struct {
	packName string
	offset   uint64
}
