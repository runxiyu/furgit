package furgit

import (
	"testing"
)

func TestErrors(t *testing.T) {
	if ErrInvalidObject == nil {
		t.Error("ErrInvalidObject should not be nil")
	}
	if ErrInvalidRef == nil {
		t.Error("ErrInvalidRef should not be nil")
	}
	if ErrNotFound == nil {
		t.Error("ErrNotFound should not be nil")
	}
}
