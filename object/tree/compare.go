package tree

// nameCompare compares two tree entry names using Git tree ordering rules.
//
// Git orders entries by name,
// treating directory names as if they carried a trailing '/'.
// entryIsTree and searchIsTree indicate
// whether the respective names belong to subtree entries.
func nameCompare(entryName string, entryIsTree bool, searchName string, searchIsTree bool) int {
	entryLen := len(entryName)
	if entryIsTree {
		entryLen++
	}

	searchLen := len(searchName)
	if searchIsTree {
		searchLen++
	}

	n := min(entryLen, searchLen)

	for i := range n {
		ec := byte('/')
		if i < len(entryName) {
			ec = entryName[i]
		}

		sc := byte('/')
		if i < len(searchName) {
			sc = searchName[i]
		}

		if ec < sc {
			return -1
		}

		if ec > sc {
			return 1
		}
	}

	switch {
	case entryLen < searchLen:
		return -1
	case entryLen > searchLen:
		return 1
	default:
		return 0
	}
}
