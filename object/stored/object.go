package stored

// Object returns the wrapped object as itself.
func (stored *Stored[T]) Object() T {
	return stored.obj
}
