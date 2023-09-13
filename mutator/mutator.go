package mutator

import (
	"math/rand"
	"strings"
	"time"
)

func NewMutator(inp string, runs uint, fnName string, validInputs ...string) *Mutator {
	return &Mutator{
		fuzzIdx:        strings.Index(inp, "FUZZ"),
		baseURL:        inp,
		input:          inp,
		fnName:         fnName,
		ch:             make(chan *Mutated, 100),
		r:              rand.New(rand.NewSource(time.Now().UnixNano())),
		runs:           runs,
		validInputs:    validInputs,
		multipleRounds: false,
	}
}

type Mutator struct {
	fuzzIdx        int
	runs           uint
	baseURL        string
	input          string
	lastInput      string
	fnName         string
	ch             chan *Mutated
	r              *rand.Rand
	validInputs    []string
	multipleRounds bool
}

type Mutated struct {
	Input    string
	Mutation string
}

func (m *Mutator) Mutate() <-chan *Mutated {
	go func() {
		if m.runs > 0 {
			for i := 0; i < int(m.runs); i++ {
				var mutatedInput string
				var method string
				mut := m.r.Intn(len(mutations) + 1)
				// run random mutations random number of times
				if mut == len(mutations) {
					m.multipleRounds = true
					countOfIterations := m.r.Intn(255)
					for iter := 0; iter < countOfIterations; iter++ {
						randomMut := m.r.Intn(len(mutations))
						ct := 0
						for k := range mutations {
							if randomMut == ct {
								mutatedInput = mutations[k](m)
								method = k
								break
							}
							ct++
						}
					}
					method = "multiple"
					m.multipleRounds = false
				} else {
					ct := 0
					for k := range mutations {
						if mut == ct {
							mutatedInput = mutations[k](m)
							method = k
							break
						}
						ct++
					}
				}

				if _, ok := applyFunctions[m.fnName]; ok {
					mutatedInput = applyFunctions[m.fnName](mutatedInput)
				}

				if m.fuzzIdx == -1 || len(m.validInputs) == 0 {
					m.ch <- &Mutated{
						Input:    mutatedInput,
						Mutation: method,
					}
				} else {
					m.ch <- &Mutated{
						Input:    strings.Replace(m.baseURL, "FUZZ", mutatedInput, -1),
						Mutation: method,
					}
				}
				m.lastInput = mutatedInput
			}
			close(m.ch)
		} else {
			for {
				var mutatedInput string
				var method string
				mut := m.r.Intn(len(mutations) + 1)
				if mut == len(mutations) {
					m.multipleRounds = true
					countOfIterations := m.r.Intn(255)
					for iter := 0; iter < countOfIterations; iter++ {
						randomMut := m.r.Intn(len(mutations))
						ct := 0
						for k := range mutations {
							if randomMut == ct {
								mutatedInput = mutations[k](m)
								method = k
								break
							}
							ct++
						}
					}
					method = "multiple"
					m.multipleRounds = false
				} else {
					ct := 0
					for k := range mutations {
						if mut == ct {
							mutatedInput = mutations[k](m)
							method = k
							break
						}
						ct++
					}

					if _, ok := applyFunctions[m.fnName]; ok {
						mutatedInput = applyFunctions[m.fnName](mutatedInput)
					}

					if m.fuzzIdx == -1 || len(m.validInputs) == 0 {
						m.ch <- &Mutated{
							Input:    mutatedInput,
							Mutation: method,
						}
					} else {
						m.ch <- &Mutated{
							Input:    strings.Replace(m.baseURL, "FUZZ", mutatedInput, -1),
							Mutation: method,
						}
					}
				}
				m.lastInput = mutatedInput
			}
		}
	}()
	return m.ch
}
