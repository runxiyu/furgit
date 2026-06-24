package clock

// get returns the value for key and marks it referenced.
func (shard *shard[K, V]) get(key K) (V, bool) {
	e, ok := shard.items.Load(key)
	if !ok {
		var zero V

		return zero, false
	}

	if !e.referenced.Load() {
		e.referenced.Store(true)
	}

	return e.value, true
}

// peek returns the value for key without affecting eviction.
func (shard *shard[K, V]) peek(key K) (V, bool) {
	e, ok := shard.items.Load(key)
	if !ok {
		var zero V

		return zero, false
	}

	return e.value, true
}
