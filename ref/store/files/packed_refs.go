package files

import (
	"lindenii.org/go/furgit/ref"
)

type packedRefs struct {
	byName  map[string]ref.Detached
	ordered []ref.Detached
}
