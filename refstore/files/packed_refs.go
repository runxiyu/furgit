package files

import (
	"codeberg.org/lindenii/furgit/ref"
)

type packedRefs struct {
	byName  map[string]ref.Detached
	ordered []ref.Detached
}
