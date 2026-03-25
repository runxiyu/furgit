package tree

// TreeEntryNameCompare compares names using Git tree ordering rules.
func TreeEntryNameCompare(entryName []byte, entryMode FileMode, searchName []byte, searchIsTree bool) int {
	isEntryTree := entryMode == FileModeDir

	entryLen := len(entryName)
	if isEntryTree {
		entryLen++
	}

	searchLen := len(searchName)
	if searchIsTree {
		searchLen++
	}

	n := min(searchLen, entryLen)

	for i := range n {
		var ec, sc byte
		if i < len(entryName) {
			ec = entryName[i]
		} else {
			ec = '/'
		}

		if i < len(searchName) {
			sc = searchName[i]
		} else {
			sc = '/'
		}

		if ec < sc {
			return -1
		}

		if ec > sc {
			return 1
		}
	}

	if entryLen < searchLen {
		return -1
	}

	if entryLen > searchLen {
		return 1
	}

	return 0
}
