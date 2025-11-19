//go:build amd64 && !purego

package adler32

import "golang.org/x/sys/cpu"

var updateFn func(d digest, p []byte) digest = updateGeneric

func init() {
	switch {
	case cpu.X86.HasAVX2:
		updateFn = updateAVX2
	default:
		updateFn = updateGeneric
	}
}

func update(d digest, p []byte) digest {
	return updateFn(d, p)
}

func updateGeneric(d digest, p []byte) digest {
	s1, s2 := uint32(d&0xffff), uint32(d>>16)

	for len(p) > 0 {
		var q []byte
		if len(p) > nmax {
			p, q = p[:nmax], p[nmax:]
		}

		for len(p) >= 32 {
			v := p[:32]
			p = p[32:]

			s1 += uint32(v[0])
			s2 += s1
			s1 += uint32(v[1])
			s2 += s1
			s1 += uint32(v[2])
			s2 += s1
			s1 += uint32(v[3])
			s2 += s1
			s1 += uint32(v[4])
			s2 += s1
			s1 += uint32(v[5])
			s2 += s1
			s1 += uint32(v[6])
			s2 += s1
			s1 += uint32(v[7])
			s2 += s1
			s1 += uint32(v[8])
			s2 += s1
			s1 += uint32(v[9])
			s2 += s1
			s1 += uint32(v[10])
			s2 += s1
			s1 += uint32(v[11])
			s2 += s1
			s1 += uint32(v[12])
			s2 += s1
			s1 += uint32(v[13])
			s2 += s1
			s1 += uint32(v[14])
			s2 += s1
			s1 += uint32(v[15])
			s2 += s1
			s1 += uint32(v[16])
			s2 += s1
			s1 += uint32(v[17])
			s2 += s1
			s1 += uint32(v[18])
			s2 += s1
			s1 += uint32(v[19])
			s2 += s1
			s1 += uint32(v[20])
			s2 += s1
			s1 += uint32(v[21])
			s2 += s1
			s1 += uint32(v[22])
			s2 += s1
			s1 += uint32(v[23])
			s2 += s1
			s1 += uint32(v[24])
			s2 += s1
			s1 += uint32(v[25])
			s2 += s1
			s1 += uint32(v[26])
			s2 += s1
			s1 += uint32(v[27])
			s2 += s1
			s1 += uint32(v[28])
			s2 += s1
			s1 += uint32(v[29])
			s2 += s1
			s1 += uint32(v[30])
			s2 += s1
			s1 += uint32(v[31])
			s2 += s1
		}

		for i := 0; i < len(p); i++ {
			x := p[i]
			s1 += uint32(x)
			s2 += s1
		}

		s1 %= mod
		s2 %= mod
		p = q
	}

	return digest(s2<<16 | s1)
}

//go:noescape
func adler32AVX2(state uint32, b []byte) uint32

func updateAVX2(d digest, p []byte) digest {
	s := uint32(d)
	for len(p) > 0 {
		chunk := p
		if len(chunk) > nmax {
			chunk = p[:nmax]
		}
		s = adler32AVX2(s, chunk)

		p = p[len(chunk):]
	}
	return digest(s)
}
