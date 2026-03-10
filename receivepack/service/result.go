package service

import (
	"codeberg.org/lindenii/furgit/packfile/ingest"
)

// Result is one receive-pack execution result.
type Result struct {
	UnpackError string
	Commands    []CommandResult
	Ingest      *ingest.Result
	Planned     []PlannedUpdate
	Applied     bool
}
