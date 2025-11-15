//go:build !sha1

package furgit

import (
	"crypto/sha256"
)

const HashSize = sha256.Size

var newHash = sha256.Sum256
