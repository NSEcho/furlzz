package mutator

import (
	"log"
	"testing"
)

func TestHandleNewCoverage(t *testing.T) {
	initialCorpus := map[string][]string{
		"FUZZ1": {"test"},
		"FUZZ2": {"123"},
	}

	m := NewMutator(
		"user=FUZZ1 pass=FUZZ2",
		"myapp",
		11,
		"url",
		false,
		initialCorpus,
	)
	defer m.Close()

	ch := m.Mutate()

	count := 0
	for {
		select {
		case mutated := <-ch:
			if mutated == nil {
				return
			}

			t.Log("Mutated Input:", mutated.Input)
			if count%5 == 0 {
				// simulate has new coverage path
				log.Println("Simulate has new coverage path")
				m.HandleNewCoverage(mutated.MutatedInputs)
				log.Println("corpus", m.inputSets)
			}
			count++
		case <-m.quit:
			return
		}
	}
}
