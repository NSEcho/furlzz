package mutator

import (
	"strings"
	"unicode"
)

type mutateFn func(m *Mutator) string

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
func insert(m *Mutator) string {
	inp := m.getFuzzedInput()
	pos := m.r.Intn(len(inp))
	var char byte
	for {
		c := m.r.Intn(unicode.MaxASCII)
		if unicode.IsPrint(rune(c)) {
			char = byte(c)
			break
		}
	}

	return inp[:pos] + string(char) + inp[pos:]
}

// del deletes random byte
func del(m *Mutator) string {
	inp := m.getFuzzedInput()
	pos := m.r.Intn(len(inp))
	return inp[:pos] + inp[pos+1:]
}

// substitute substitutes byte at random position with random byte
func substitute(m *Mutator) string {
	inp := m.getFuzzedInput()
	pos := m.r.Intn(len(inp))
	var char byte
	for {
		c := m.r.Intn(unicode.MaxASCII)
		if unicode.IsPrint(rune(c)) {
			char = byte(c)
			break
		}
	}
	var res string
	for i, c := range inp {
		if i == pos {
			res += string(char)
		} else {
			res += string(c)
		}
	}
	return res
}

// byteOp takes random byte and random position inside the string
// and do arithmetic operation on them (+, -, *, /)
func byteOp(m *Mutator) string {
	b := make([]byte, 1)
	m.r.Read(b)
	inp := m.getFuzzedInput()
	pos := m.r.Intn(len(inp))

	op := m.r.Intn(4)

	res := make([]rune, len(inp))
	for i, r := range inp {
		if i == pos {
			switch op {
			case 0:
				res[i] = r + rune(b[0])
			case 1:
				res[i] = r - rune(b[0])
			case 2:
				res[i] = r * rune(b[0])
			default:
				if b[0] != 0 {
					res[i] = r / rune(b[0])
				} else {
					res[i] = r + rune(b[0])
				}
			}
		} else {
			res[i] = r
		}
	}

	return string(res)
}

// duplicateRange duplicates random range inside the original string random
// number of times
func duplicateRange(m *Mutator) string {
	inp := m.getFuzzedInput()

	start := m.r.Intn(len(inp))

	var end int
	for end = m.r.Intn(len(inp)); end < start; end = m.r.Intn(len(inp)) {
	}

	var countOfDuplications int
	for countOfDuplications = m.r.Intn(len(inp)); countOfDuplications < 1; countOfDuplications = m.r.Intn(len(inp)) {
	}

	rng := inp[start:end]

	res := ""
	res += inp[:start]
	res += strings.Repeat(rng, countOfDuplications)
	res += inp[end:]

	return res
}

// bitFlip flips the bit at random position inside random location inside input
func bitFlip(m *Mutator) string {
	inp := m.getFuzzedInput()

	pos := m.r.Intn(len(inp))
	bitPosition := m.r.Intn(8)

	res := ""

	for i, r := range inp {
		if i == pos {
			res += string(r ^ (1 << bitPosition))
		} else {
			res += string(r)
		}
	}

	return res
}

// bitmask applies random bitmask on random location inside the string
func bitmask(m *Mutator) string {
	inp := m.getFuzzedInput()

	pos := m.r.Intn(len(inp))
	bm := m.r.Intn(255)

	res := ""

	for i, r := range inp {
		if pos == i {
			res += string(inp[i] ^ uint8(bm))
		} else {
			res += string(r)
		}
	}

	return res
}

// duplicate duplicates original string random number of times (2 < 10)
func duplicate(m *Mutator) string {
	inp := m.getFuzzedInput()

	var count int
	for count = m.r.Intn(10); count < 1; count = m.r.Intn(10) {
	}

	return strings.Repeat(inp, count)
}
