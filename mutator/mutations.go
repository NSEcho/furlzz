package mutator

import (
	"bytes"
)

type mutateFn func(m *Mutator) []byte

var mutations = map[string]mutateFn{
	"insert":          insert,
	"delete":          del,
	"substitute":      substitute,
	"byte op":         byteOp,
	"duplicate range": duplicateRange,
	"bit flip":        bitFlip,
	"bitmask":         bitmask,
	"duplicate":       duplicate,
}

// insert inserts random byte at random location inside the input
func insert(m *Mutator) []byte {
	inp := m.getFuzzedInput()
	pos := m.r.Intn(len(inp))
	char := byte(m.r.Intn(255))

	res := make([]byte, len(inp)+1)

	k := 0

	for i := 0; i < len(inp); i++ {
		if i == pos {
			res[k] = char
			res[k+1] = inp[i]
			k += 2
		} else {
			res[k] = inp[i]
			k++
		}
	}

	return res
}

// del deletes random byte
func del(m *Mutator) []byte {
	inp := m.getFuzzedInput()
	pos := m.r.Intn(len(inp))
	res := make([]byte, len(inp)-1)

	k := 0
	for i := 0; i < len(inp); i++ {
		if i == pos {
			continue
		}
		res[k] = inp[i]
		k++
	}

	return res
}

// substitute substitutes byte at random position with random byte
func substitute(m *Mutator) []byte {
	inp := m.getFuzzedInput()
	pos := m.r.Intn(len(inp))
	char := byte(m.r.Intn(255))

	res := make([]byte, len(inp))

	for i, c := range inp {
		if i == pos {
			res[i] = char
		} else {
			res[i] = c
		}
	}
	return res
}

// byteOp takes random byte and random position inside the string
// and do arithmetic operation on them (+, -, *, /)
func byteOp(m *Mutator) []byte {
	b := make([]byte, 1)
	m.r.Read(b)
	inp := m.getFuzzedInput()
	pos := m.r.Intn(len(inp))

	op := m.r.Intn(4)

	res := make([]byte, len(inp))
	for i, r := range inp {
		if i == pos {
			switch op {
			case 0:
				res[i] = r + b[0]
			case 1:
				res[i] = r - b[0]
			case 2:
				res[i] = r * b[0]
			default:
				if b[0] != 0 {
					res[i] = r / b[0]
				} else {
					res[i] = r + b[0]
				}
			}
		} else {
			res[i] = r
		}
	}

	return res
}

// duplicateRange duplicates random range inside the original string random
// number of times
func duplicateRange(m *Mutator) []byte {
	inp := m.getFuzzedInput()

	start := m.r.Intn(len(inp))

	var end int
	for end = m.r.Intn(len(inp)); end < start; end = m.r.Intn(len(inp)) {
	}

	var countOfDuplications int
	for countOfDuplications = m.r.Intn(len(inp)); countOfDuplications < 1; countOfDuplications = m.r.Intn(len(inp)) {
	}

	rng := inp[start:end]
	duplicatedBytes := bytes.Repeat(rng, countOfDuplications)

	res := make([]byte, len(inp)+len(duplicatedBytes))

	k := 0
	for i := 0; i < end; i++ {
		res[k] = inp[i]
		k++
	}

	for i := 0; i < len(duplicatedBytes); i++ {
		res[k] = duplicatedBytes[i]
		k++
	}

	for i := end; i < len(inp); i++ {
		res[k] = inp[i]
		k++
	}

	return res
}

// bitFlip flips the bit at random position inside random location inside input
func bitFlip(m *Mutator) []byte {
	inp := m.getFuzzedInput()

	pos := m.r.Intn(len(inp))
	bitPosition := m.r.Intn(8)

	res := make([]byte, len(inp))

	for i, r := range inp {
		if i == pos {
			res[i] = r ^ (1 << bitPosition)
		} else {
			res[i] = r
		}
	}

	return res
}

// bitmask applies random bitmask on random location inside the string
func bitmask(m *Mutator) []byte {
	inp := m.getFuzzedInput()

	pos := m.r.Intn(len(inp))
	bm := m.r.Intn(255)

	res := make([]byte, len(inp))

	for i, r := range inp {
		if pos == i {
			res[i] = inp[i] ^ uint8(bm)
		} else {
			res[i] = r
		}
	}

	return res
}

// duplicate duplicates original string random number of times (2 < 10)
func duplicate(m *Mutator) []byte {
	inp := m.getFuzzedInput()

	var count int
	for count = m.r.Intn(10); count < 1; count = m.r.Intn(10) {
	}

	return bytes.Repeat(inp, count)
}
