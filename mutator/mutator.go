package mutator

import (
	"bytes"
	"fmt"
	"math/rand"
	"strings"
	"time"
)

func NewMutator(inp, app string, runs uint, fnName string, ignoreCrashes bool, validInputs ...string) *Mutator {
	var crashes []string
	if ignoreCrashes {
		c, err := readCrashes(app)
		if err != nil {
			panic(fmt.Sprintf("error reading crashes: %v", err))
		}
		crashes = c
	}

	return &Mutator{
		fuzzIdx:        strings.Index(inp, "FUZZ"),
		baseURL:        inp,
		input:          inp,
		fnName:         fnName,
		ignoreCrashes:  ignoreCrashes,
		ch:             make(chan *Mutated, 100),
		r:              rand.New(rand.NewSource(time.Now().UnixNano())),
		runs:           runs,
		validInputs:    validInputs,
		crashes:        crashes,
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
	ignoreCrashes  bool
	ch             chan *Mutated
	r              *rand.Rand
	validInputs    []string
	crashes        []string
	multipleRounds bool
	quit           chan struct{}
}

type Mutated struct {
	Input    string
	Mutation string
}

func (m *Mutator) Close() {
	m.quit <- struct{}{}
	close(m.ch)
	close(m.quit)
}

func (m *Mutator) Mutate() <-chan *Mutated {
	go func() {
		if m.runs > 0 {
			for i := 0; i < int(m.runs); i++ {
				inp := m.mutateAndSend()
				for !inp {
					inp = m.mutateAndSend()
				}
			}
			close(m.ch)
		} else {
			for {
				select {
				case <-m.quit:
					break
				default:
					inp := m.mutateAndSend()
					for !inp {
						inp = m.mutateAndSend()
					}
				}
			}
		}
	}()
	return m.ch
}

func (m *Mutator) mutateAndSend() bool {
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

	if m.ignoreCrashes && len(m.crashes) > 0 {
		for _, crash := range m.crashes {
			crashInp := strings.Replace(crash, m.baseURL[:m.fuzzIdx], "", -1)
			if bytes.Equal([]byte(crashInp), []byte(mutatedInput)) {
				return false
			}
		}
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
	return true
}
