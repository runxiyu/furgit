package main

import (
	"fmt"
)

func (explainer *explainer) walkDelta(base, payload []byte, pos int) ([]byte, bool, error) {
	explainer.printf("\tdelta\n")

	building := base != nil

	var result []byte

	insn := 0

	for pos < len(payload) {
		op := payload[pos]
		pos++
		insn++

		switch {
		case op&0x80 != 0:
			next, seg, err := explainer.decodeCopy(base, payload, pos, op)
			if err != nil {
				return nil, false, err
			}

			pos = next

			if building {
				result = append(result, seg...)
			}
		case op != 0:
			next, lit, err := explainer.decodeInsert(payload, pos, int(op))
			if err != nil {
				return nil, false, err
			}

			pos = next

			if building {
				result = append(result, lit...)
			}
		default:
			explainer.printf("\t\tinvalid opcode 0x00; stopping delta decode\n")

			return nil, false, nil
		}
	}

	if !building {
		return nil, false, nil
	}

	return result, true, nil
}

func (explainer *explainer) decodeCopy(base, payload []byte, pos int, op byte) (int, []byte, error) {
	offset := 0

	for i := range 4 {
		if op&(1<<uint(i)) == 0 { //nolint:gosec
			continue
		}

		if pos >= len(payload) {
			return 0, nil, fmt.Errorf("truncated copy offset")
		}

		offset |= int(payload[pos]) << (8 * uint(i)) //nolint:gosec
		pos++
	}

	size := 0

	for i := range 3 {
		if op&(1<<uint(4+i)) == 0 { //nolint:gosec
			continue
		}

		if pos >= len(payload) {
			return 0, nil, fmt.Errorf("truncated copy size")
		}

		size |= int(payload[pos]) << (8 * uint(i)) //nolint:gosec
		pos++
	}

	if size == 0 {
		size = 0x10000
	}

	explainer.printf("\t\tcpy %d from %d\n", size, offset)

	if base == nil {
		return pos, nil, nil
	}

	if offset < 0 || offset+size > len(base) {
		return 0, nil, fmt.Errorf("copy of %d byte(s) from base offset %d exceeds base length %d", size, offset, len(base))
	}

	seg := base[offset : offset+size]
	hexBlock(explainer.out, "\t\t\t", seg)

	return pos, seg, nil
}

func (explainer *explainer) decodeInsert(payload []byte, pos, n int) (int, []byte, error) {
	if pos+n > len(payload) {
		return 0, nil, fmt.Errorf("truncated insert payload")
	}

	lit := payload[pos : pos+n]

	explainer.printf("\t\tins %d\n", n)
	hexBlock(explainer.out, "\t\t\t", lit)

	return pos + n, lit, nil
}
