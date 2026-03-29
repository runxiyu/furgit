package objectid

import "crypto/sha256"

// maxObjectIDSize MUST be >= the largest supported algorithm size.
const maxObjectIDSize = sha256.Size
