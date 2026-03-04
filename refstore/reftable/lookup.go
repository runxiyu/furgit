package reftable

import (
	"encoding/binary"
	"fmt"
	"strings"

	"codeberg.org/lindenii/furgit/internal/intconv"
	"codeberg.org/lindenii/furgit/objectid"
)

// resolveRecord resolves one ref name inside a single table file.
func (table *tableFile) resolveRecord(name string) (recordValue, bool, error) {
	if table.refIndexPos != 0 {
		indexPos, err := intconv.Uint64ToInt(table.refIndexPos)
		if err != nil {
			return recordValue{}, false, err
		}

		pos, ok, err := table.resolveRefBlockPosFromIndex(name, indexPos)
		if err != nil {
			return recordValue{}, false, err
		}

		if !ok {
			return recordValue{}, false, nil
		}

		return table.lookupInRefBlock(name, pos)
	}

	// Without a ref index, fall back to scanning ref blocks in order.
	pos := table.headerLen
	for pos < table.refEnd {
		for pos < table.refEnd && table.data[pos] == 0 {
			pos++
		}

		if pos >= table.refEnd {
			break
		}

		if table.data[pos] != blockTypeRef {
			return recordValue{}, false, fmt.Errorf("refstore/reftable: table %q: unexpected block type %q in ref section", table.name, table.data[pos])
		}

		block, blockEnd, err := table.readBlockAt(pos)
		if err != nil {
			return recordValue{}, false, err
		}

		found, done, rec, err := lookupRecordInRefBlock(table, block, name)
		if err != nil {
			return recordValue{}, false, err
		}

		if found {
			return rec, true, nil
		}

		if done {
			return recordValue{}, false, nil
		}

		pos = table.nextBlockPos(blockEnd)
	}

	return recordValue{}, false, nil
}

// resolveRefBlockPosFromIndex resolves a candidate ref block position via index blocks.
func (table *tableFile) resolveRefBlockPosFromIndex(name string, indexPos int) (int, bool, error) {
	block, _, err := table.readBlockAt(indexPos)
	if err != nil {
		return 0, false, err
	}

	if block.blockType != blockTypeIndex {
		return 0, false, fmt.Errorf("refstore/reftable: table %q: ref index root is not index block", table.name)
	}

	childPos, ok, err := lookupChildPosInIndexBlock(block, name)
	if err != nil {
		return 0, false, err
	}

	if !ok {
		return 0, false, nil
	}

	if childPos < 0 || childPos >= len(table.data) {
		return 0, false, fmt.Errorf("refstore/reftable: table %q: index child position out of range", table.name)
	}

	childType := table.data[childPos]
	switch childType {
	case blockTypeRef:
		return childPos, true, nil
	case blockTypeIndex:
		return table.resolveRefBlockPosFromIndex(name, childPos)
	default:
		return 0, false, fmt.Errorf("refstore/reftable: table %q: unexpected child block type %q", table.name, childType)
	}
}

// lookupInRefBlock searches one ref block by full ref name.
func (table *tableFile) lookupInRefBlock(name string, pos int) (recordValue, bool, error) {
	block, _, err := table.readBlockAt(pos)
	if err != nil {
		return recordValue{}, false, err
	}

	if block.blockType != blockTypeRef {
		return recordValue{}, false, fmt.Errorf("refstore/reftable: table %q: expected ref block at %d", table.name, pos)
	}

	found, _, rec, err := lookupRecordInRefBlock(table, block, name)
	if err != nil {
		return recordValue{}, false, err
	}

	return rec, found, nil
}

// forEachRecord iterates all ref records in this table in lexical order.
func (table *tableFile) forEachRecord(fn func(name string, rec recordValue) error) error {
	pos := table.headerLen
	prevLast := ""

	for pos < table.refEnd {
		for pos < table.refEnd && table.data[pos] == 0 {
			pos++
		}

		if pos >= table.refEnd {
			break
		}

		if table.data[pos] != blockTypeRef {
			return fmt.Errorf("refstore/reftable: table %q: unexpected block type %q in ref section", table.name, table.data[pos])
		}

		block, blockEnd, err := table.readBlockAt(pos)
		if err != nil {
			return err
		}

		var first, last string

		err = forEachRecordInRefBlock(table, block, func(name string, rec recordValue) error {
			if first == "" {
				first = name
			}

			last = name

			return fn(name, rec)
		})
		if err != nil {
			return err
		}

		if prevLast != "" && first != "" && strings.Compare(first, prevLast) <= 0 {
			return fmt.Errorf("refstore/reftable: table %q: ref blocks are not strictly ordered", table.name)
		}

		if last != "" {
			prevLast = last
		}

		pos = table.nextBlockPos(blockEnd)
	}

	return nil
}

// blockView is one decoded block boundary within the mapped table bytes.
type blockView struct {
	blockType byte
	start     int
	end       int
	first     bool
	payload   []byte
}

// readBlockAt validates and returns a block view starting at pos.
func (table *tableFile) readBlockAt(pos int) (blockView, int, error) {
	if pos < 0 || pos+4 > len(table.data) {
		return blockView{}, 0, fmt.Errorf("refstore/reftable: table %q: block header out of range", table.name)
	}

	blockLen := int(readUint24(table.data[pos+1 : pos+4]))

	effectiveLen := blockLen
	if pos == table.headerLen {
		if blockLen < table.headerLen {
			return blockView{}, 0, fmt.Errorf("refstore/reftable: table %q: invalid first block length", table.name)
		}

		effectiveLen = blockLen - table.headerLen
	}

	if effectiveLen < 4 {
		return blockView{}, 0, fmt.Errorf("refstore/reftable: table %q: invalid block length", table.name)
	}

	end := pos + effectiveLen
	if end > len(table.data) {
		return blockView{}, 0, fmt.Errorf("refstore/reftable: table %q: block out of range", table.name)
	}

	view := blockView{blockType: table.data[pos], start: pos, end: end, first: pos == table.headerLen, payload: table.data[pos:end]}

	return view, end, nil
}

// nextBlockPos computes the next block start from current block end.
func (table *tableFile) nextBlockPos(blockEnd int) int {
	if table.blockSize > 0 {
		return alignUp(blockEnd, table.blockSize)
	}

	return blockEnd
}

// lookupChildPosInIndexBlock selects a child block position for key.
func lookupChildPosInIndexBlock(block blockView, key string) (int, bool, error) {
	off, recordsEnd, restarts, err := parseBlockLayout(block)
	if err != nil {
		return 0, false, err
	}

	err = validateRestarts(block, restarts, off, recordsEnd, true)
	if err != nil {
		return 0, false, err
	}

	prev := ""
	for off < recordsEnd {
		name, v, nextOff, err := parseKeyedRecord(block.payload, off, recordsEnd, prev)
		if err != nil {
			return 0, false, err
		}

		if (v & 0x7) != 0 {
			return 0, false, fmt.Errorf("index value_type must be 0")
		}

		childPos, nextOff, err := readVarint(block.payload, nextOff, recordsEnd)
		if err != nil {
			return 0, false, err
		}

		if strings.Compare(key, name) <= 0 {
			childPosInt, err := intconv.Uint64ToInt(childPos)
			if err != nil {
				return 0, false, fmt.Errorf("index child position conversion: %w", err)
			}

			return childPosInt, true, nil
		}

		prev = name
		off = nextOff
	}

	if off != recordsEnd {
		return 0, false, fmt.Errorf("malformed index block")
	}

	return 0, false, nil
}

// lookupRecordInRefBlock searches one ref block and may short-circuit by sort order.
func lookupRecordInRefBlock(table *tableFile, block blockView, key string) (found, done bool, rec recordValue, err error) {
	off, recordsEnd, restarts, err := parseBlockLayout(block)
	if err != nil {
		return false, false, recordValue{}, err
	}

	err = validateRestarts(block, restarts, off, recordsEnd, true)
	if err != nil {
		return false, false, recordValue{}, err
	}

	prev := ""
	for off < recordsEnd {
		name, v, nextOff, err := parseKeyedRecord(block.payload, off, recordsEnd, prev)
		if err != nil {
			return false, false, recordValue{}, err
		}

		typeBits := byte(v & 0x7)

		_, nextOff, err = readVarint(block.payload, nextOff, recordsEnd)
		if err != nil {
			return false, false, recordValue{}, err
		}

		recVal, nextOff, err := parseRefValue(block.payload, nextOff, recordsEnd, table.algo, typeBits)
		if err != nil {
			return false, false, recordValue{}, err
		}

		cmp := strings.Compare(name, key)
		if cmp == 0 {
			return true, true, recVal, nil
		}

		if cmp > 0 {
			return false, true, recordValue{}, nil
		}

		prev = name
		off = nextOff
	}

	if off != recordsEnd {
		return false, false, recordValue{}, fmt.Errorf("malformed ref block")
	}

	return false, false, recordValue{}, nil
}

// forEachRecordInRefBlock iterates all records in one ref block.
func forEachRecordInRefBlock(table *tableFile, block blockView, fn func(name string, rec recordValue) error) error {
	off, recordsEnd, restarts, err := parseBlockLayout(block)
	if err != nil {
		return err
	}

	err = validateRestarts(block, restarts, off, recordsEnd, true)
	if err != nil {
		return err
	}

	prev := ""
	for off < recordsEnd {
		name, v, nextOff, err := parseKeyedRecord(block.payload, off, recordsEnd, prev)
		if err != nil {
			return err
		}

		typeBits := byte(v & 0x7)

		_, nextOff, err = readVarint(block.payload, nextOff, recordsEnd)
		if err != nil {
			return err
		}

		recVal, nextOff, err := parseRefValue(block.payload, nextOff, recordsEnd, table.algo, typeBits)
		if err != nil {
			return err
		}

		err = fn(name, recVal)
		if err != nil {
			return err
		}

		prev = name
		off = nextOff
	}

	if off != recordsEnd {
		return fmt.Errorf("malformed ref block")
	}

	return nil
}

// parseBlockLayout parses common record/restart regions for ref and index blocks.
func parseBlockLayout(block blockView) (recordsStart, recordsEnd int, restarts []int, err error) {
	if len(block.payload) < 6 {
		return 0, 0, nil, fmt.Errorf("short block")
	}

	restartCount := int(binary.BigEndian.Uint16(block.payload[len(block.payload)-2:]))
	if restartCount <= 0 {
		return 0, 0, nil, fmt.Errorf("invalid restart count")
	}

	restarts = make([]int, restartCount)
	restartBytes := restartCount * 3

	restartsStart := len(block.payload) - 2 - restartBytes
	if restartsStart < 4 {
		return 0, 0, nil, fmt.Errorf("invalid restart table")
	}

	for i := range restartCount {
		off := restartsStart + i*3
		rel := int(readUint24(block.payload[off : off+3]))

		base := block.start
		if block.first {
			// In the first block, restart offsets are relative to file start.
			base = 0
		}

		abs := base + rel
		restarts[i] = abs - block.start
	}

	return 4, restartsStart, restarts, nil
}

// validateRestarts validates restart monotonicity, bounds and record-prefix invariants.
func validateRestarts(block blockView, restarts []int, recordsStart, recordsEnd int, requirePrefixZero bool) error {
	prev := -1

	for _, off := range restarts {
		if off < recordsStart || off >= recordsEnd {
			return fmt.Errorf("restart offset out of range")
		}

		if off <= prev {
			return fmt.Errorf("restart offsets not strictly increasing")
		}

		prev = off
		if requirePrefixZero {
			prefix, _, err := readVarint(block.payload, off, recordsEnd)
			if err != nil {
				return err
			}

			if prefix != 0 {
				return fmt.Errorf("restart record prefix length must be zero")
			}
		}
	}

	return nil
}

// parseKeyedRecord parses one prefix-compressed key record header.
func parseKeyedRecord(buf []byte, off, end int, prev string) (name string, rawType uint64, next int, err error) {
	prefixLen, next, err := readVarint(buf, off, end)
	if err != nil {
		return "", 0, 0, err
	}

	suffixAndType, next, err := readVarint(buf, next, end)
	if err != nil {
		return "", 0, 0, err
	}

	suffixLen, err := intconv.Uint64ToInt(suffixAndType >> 3)
	if err != nil || suffixLen < 0 || next+suffixLen > end {
		return "", 0, 0, fmt.Errorf("invalid suffix length")
	}

	prefixLenInt, err := intconv.Uint64ToInt(prefixLen)
	if err != nil {
		return "", 0, 0, fmt.Errorf("invalid prefix length")
	}

	if prefixLenInt > len(prev) {
		return "", 0, 0, fmt.Errorf("invalid prefix length")
	}

	name = prev[:prefixLenInt] + string(buf[next:next+suffixLen])
	next += suffixLen

	if prev != "" && strings.Compare(name, prev) <= 0 {
		return "", 0, 0, fmt.Errorf("keys not strictly increasing")
	}

	return name, suffixAndType, next, nil
}

// parseRefValue parses one ref-record value payload according to value_type.
func parseRefValue(buf []byte, off, end int, algo objectid.Algorithm, valueType byte) (recordValue, int, error) {
	switch valueType {
	case 0x0:
		return recordValue{deleted: true}, off, nil
	case 0x1:
		id, next, err := readObjectID(buf, off, end, algo)
		if err != nil {
			return recordValue{}, 0, err
		}

		return recordValue{detachedID: id, hasDetached: true}, next, nil
	case 0x2:
		id, next, err := readObjectID(buf, off, end, algo)
		if err != nil {
			return recordValue{}, 0, err
		}

		peeled, next, err := readObjectID(buf, next, end, algo)
		if err != nil {
			return recordValue{}, 0, err
		}

		peeledCopy := peeled

		return recordValue{detachedID: id, hasDetached: true, peeled: &peeledCopy}, next, nil
	case 0x3:
		targetLen, next, err := readVarint(buf, off, end)
		if err != nil {
			return recordValue{}, 0, err
		}

		remaining := end - next
		if remaining < 0 {
			return recordValue{}, 0, fmt.Errorf("invalid symref target length")
		}

		remainingU64, err := intconv.IntToUint64(remaining)
		if err != nil {
			return recordValue{}, 0, fmt.Errorf("invalid symref target length")
		}

		if targetLen > remainingU64 {
			return recordValue{}, 0, fmt.Errorf("invalid symref target length")
		}

		targetLenInt, err := intconv.Uint64ToInt(targetLen)
		if err != nil {
			return recordValue{}, 0, fmt.Errorf("invalid symref target length")
		}

		target := string(buf[next : next+targetLenInt])
		next += targetLenInt

		return recordValue{symbolicTarget: target}, next, nil
	default:
		return recordValue{}, 0, fmt.Errorf("unsupported ref value type %d", valueType)
	}
}

// readObjectID reads one object ID using the table algorithm width.
func readObjectID(buf []byte, off, end int, algo objectid.Algorithm) (objectid.ObjectID, int, error) {
	sz := algo.Size()
	if off < 0 || sz < 0 || off+sz > end {
		return objectid.ObjectID{}, 0, fmt.Errorf("truncated object id")
	}

	id, err := objectid.FromBytes(algo, buf[off:off+sz])
	if err != nil {
		return objectid.ObjectID{}, 0, err
	}

	return id, off + sz, nil
}
