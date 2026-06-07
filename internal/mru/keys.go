package mru

// Keys returns the keys in most-recently-used order,
// front first.
//
// The result is the immutable snapshot current at the call:
// a concurrent Touch or Sync does not affect it.
//
// Labels: Mut-No.
func (order *Order[K]) Keys() []K {
	keys := order.snapshot.Load()
	if keys == nil {
		return nil
	}

	return *keys
}
