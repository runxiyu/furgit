package furgit

const deltaCopyDefaultLen = 0x10000

func deltaEncode(baseSize, resultSize int, instr []deltaInstruction) ([]byte, error) {
	if baseSize < 0 || resultSize < 0 {
		return nil, ErrInvalidObject
	}
	out := make([]byte, 0, 64)
	baseHdr, err := packVarintEncode(baseSize)
	if err != nil {
		return nil, err
	}
	out = append(out, baseHdr...)
	resultHdr, err := packVarintEncode(resultSize)
	if err != nil {
		return nil, err
	}
	out = append(out, resultHdr...)

	for _, ins := range instr {
		if ins.copy {
			if ins.offset < 0 || ins.length <= 0 {
				return nil, ErrInvalidObject
			}
			if ins.offset > 0xffffffff {
				return nil, ErrInvalidObject
			}
			if ins.length > 0xffffff {
				return nil, ErrInvalidObject
			}
			op := byte(0x80)
			var tmp [7]byte
			pos := 0

			off := ins.offset
			for i := 0; i < 4; i++ {
				op |= 1 << i
				tmp[pos] = byte(off & 0xff)
				pos++
				off >>= 8
				if off == 0 {
					break
				}
			}
			if off != 0 {
				return nil, ErrInvalidObject
			}

			n := ins.length
			if n != deltaCopyDefaultLen {
				for i := 0; i < 3 && n > 0; i++ {
					op |= 1 << (i + 4)
					tmp[pos] = byte(n & 0xff)
					pos++
					n >>= 8
				}
				if n > 0 {
					return nil, ErrInvalidObject
				}
			}

			out = append(out, op)
			out = append(out, tmp[:pos]...)
			continue
		}

		data := ins.data
		if ins.length > 0 {
			if len(data) < ins.length {
				return nil, ErrInvalidObject
			}
			data = data[:ins.length]
		}
		if len(data) == 0 {
			return nil, ErrInvalidObject
		}
		for len(data) > 0 {
			n := len(data)
			if n > 127 {
				n = 127
			}
			out = append(out, byte(n))
			out = append(out, data[:n]...)
			data = data[n:]
		}
	}

	return out, nil
}

func deltaTry(base, target []byte, seed uint32, minSavings int) ([]byte, bool) {
	if minSavings < 0 {
		minSavings = 0
	}
	instr, err := deltify(base, target, seed)
	if err != nil {
		return nil, false
	}
	delta, err := deltaEncode(len(base), len(target), instr)
	if err != nil {
		return nil, false
	}
	if len(delta)+minSavings >= len(target) {
		return nil, false
	}
	return delta, true
}
