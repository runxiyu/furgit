package lru

type entry[K comparable, V any] struct {
	key    K
	value  V
	weight int64
}
