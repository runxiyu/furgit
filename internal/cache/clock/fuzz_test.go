package clock //nolint:testpackage

import "testing"

// FuzzShard replays a decoded op stream against one shard,
// checking the value oracle and invariants after every op.
func FuzzShard(f *testing.F) {
	f.Add([]byte{})
	f.Add([]byte{0, 1, 10, 0, 2, 10, 0, 3, 10, 0, 4, 10, 1, 1, 0, 0, 5, 10})
	f.Add([]byte{0, 7, 200, 0, 7, 5, 2, 7, 0, 3, 0, 0, 0, 8, 8})

	f.Fuzz(func(t *testing.T, program []byte) {
		const maxWeight = 32

		shard := newShard[uint8, uint64](maxWeight)
		shadow := make(map[uint8]uint64)

		var nonce uint64

		for i := 0; i+2 < len(program); i += 3 {
			key := program[i+1]
			weight := uint64(program[i+2])

			switch program[i] % 4 {
			case 0: // add
				nonce++
				value := nonce

				admitted := shard.add(key, value, weight)
				if admitted != (weight <= maxWeight) {
					t.Fatalf("add(%d, w=%d) admitted=%v, want %v", key, weight, admitted, weight <= maxWeight)
				}

				if admitted {
					shadow[key] = value
				}
			case 1: // get
				if got, ok := shard.get(key); ok && got != shadow[key] {
					t.Fatalf("get(%d) = %d, want %d", key, got, shadow[key])
				}
			case 2: // peek
				if got, ok := shard.peek(key); ok && got != shadow[key] {
					t.Fatalf("peek(%d) = %d, want %d", key, got, shadow[key])
				}
			case 3: // clear
				shard.clear()
				clear(shadow)
			}

			checkShard(t, shard)
		}
	})
}
