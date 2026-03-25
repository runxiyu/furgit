package server

import (
	"slices"
)

func sortAdvertisedRefs(refs []AdvertisedRef) []AdvertisedRef {
	out := append([]AdvertisedRef(nil), refs...)
	slices.SortFunc(out, func(left, right AdvertisedRef) int {
		if left.Name == "HEAD" && right.Name != "HEAD" {
			return -1
		}

		if left.Name != "HEAD" && right.Name == "HEAD" {
			return 1
		}

		switch {
		case left.Name < right.Name:
			return -1
		case left.Name > right.Name:
			return 1
		default:
			return 0
		}
	})

	return out
}
