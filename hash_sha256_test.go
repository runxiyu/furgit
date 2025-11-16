//go:build !sha1

package furgit

import (
	"crypto/sha256"
)

const testHashSize = sha256.Size
