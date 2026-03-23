//go:build amd64 && !purego

package cpu

const (
	cpuidOSXSAVE = 1 << 27
	cpuidAVX     = 1 << 28
	cpuidAVX2    = 1 << 5
)

// cpuid is implemented in cpu_amd64.s.
func cpuid(eaxArg, ecxArg uint32) (eax, ebx, ecx, edx uint32)

// xgetbv with ecx = 0 is implemented in cpu_amd64.s.
func xgetbv() (eax, edx uint32)

func init() {
	maxID, _, _, _ := cpuid(0, 0)
	if maxID < 7 {
		return
	}

	_, _, ecx1, _ := cpuid(1, 0)

	osSupportsAVX := false
	if ecx1&cpuidOSXSAVE != 0 {
		eax, _ := xgetbv()
		osSupportsAVX = eax&(1<<1) != 0 && eax&(1<<2) != 0
	}

	_, ebx7, _, _ := cpuid(7, 0)
	X86.HasAVX2 = osSupportsAVX && ecx1&cpuidAVX != 0 && ebx7&cpuidAVX2 != 0
}
