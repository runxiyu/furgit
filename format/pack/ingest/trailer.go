package ingest

import (
	"bytes"
	"fmt"
	"hash"
)

// trailerVerifier incrementally verifies trailing hash bytes in a stream.
type trailerVerifier struct {
	hash     hash.Hash
	hashSize int
	tail     []byte
	seen     int64
}

// newTrailerVerifier creates a trailing hash verifier.
func newTrailerVerifier(hash hash.Hash, hashSize int) *trailerVerifier {
	return &trailerVerifier{
		hash:     hash,
		hashSize: hashSize,
		tail:     make([]byte, 0, hashSize),
	}
}

// write feeds one chunk of stream bytes into the verifier.
func (verifier *trailerVerifier) write(src []byte) {
	if len(src) == 0 {
		return
	}

	verifier.seen += int64(len(src))
	if len(verifier.tail) == 0 && len(src) <= verifier.hashSize {
		verifier.tail = append(verifier.tail, src...)

		return
	}

	tmp := make([]byte, 0, len(verifier.tail)+len(src))
	tmp = append(tmp, verifier.tail...)
	tmp = append(tmp, src...)
	if len(tmp) <= verifier.hashSize {
		verifier.tail = tmp

		return
	}

	flushN := len(tmp) - verifier.hashSize
	_, _ = verifier.hash.Write(tmp[:flushN])
	verifier.tail = append(verifier.tail[:0], tmp[flushN:]...)
}

// verify finalizes verification against the stream trailer.
func (verifier *trailerVerifier) verify() error {
	if len(verifier.tail) != verifier.hashSize {
		return fmt.Errorf("format/pack/ingest: stream too short for trailer hash")
	}

	computed := verifier.hash.Sum(nil)
	if !bytes.Equal(computed, verifier.tail) {
		return &ErrPackTrailerMismatch{}
	}

	return nil
}
