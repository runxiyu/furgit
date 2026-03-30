package service

// Result is one receive-pack execution result.
type Result struct {
	UnpackError string
	Commands    []CommandResult
	Planned     []PlannedUpdate
	Applied     bool
}
