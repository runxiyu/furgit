package service

import (
	"codeberg.org/lindenii/furgit/format/pack/ingest"
)

// Result is one receive-pack execution result.
type Result struct {
	UnpackError string
	Commands    []CommandResult
	Ingest      *ingest.Result
	Planned     []PlannedUpdate
	Applied     bool
}
