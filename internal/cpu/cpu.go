package cpu

// X86 contains x86 CPU feature flags detected at runtime.
//
//nolint:gochecknoglobals
var X86 struct {
	HasAVX2 bool
}
