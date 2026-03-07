package server_test

import (
	"bytes"
	"strings"
	"testing"

	"codeberg.org/lindenii/furgit/objectid"
)

type bufferWriteFlusher struct {
	bytes.Buffer
}

func (bufferWriteFlusher) Flush() error {
	return nil
}

func mustHexID(tb testing.TB, algo objectid.Algorithm, digit string) objectid.ObjectID {
	tb.Helper()

	id, err := objectid.ParseHex(algo, strings.Repeat(digit, algo.HexLen()))
	if err != nil {
		tb.Fatalf("objectid.ParseHex(%q): %v", strings.Repeat(digit, algo.HexLen()), err)
	}

	return id
}
