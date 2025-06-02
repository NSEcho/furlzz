package mutator

import (
	"io"
	"os"
	"path/filepath"
)

func (m *Mutator) getFuzzedInput(set string) string {
	if m.multipleRounds {
		if m.lastInput == "" {
			m.lastInput = m.fetchInput(set)
		}
		return m.lastInput
	}
	return m.fetchInput(set)
}

func (m *Mutator) fetchInput(set string) string {
	k := m.r.Intn(len(m.inputSets[set]))
	return m.inputSets[set][k]
}

func readCrashes(app string) ([]string, error) {
	files, _ := filepath.Glob("fcrash_*_*")

	var crashes []string
	for _, fl := range files {
		data, err := func() ([]byte, error) {
			f, err := os.Open(fl)
			if err != nil {
				return nil, err
			}
			defer f.Close()

			data, _ := io.ReadAll(f)
			return data, nil
		}()
		if err != nil {
			return nil, err
		}
		crashes = append(crashes, string(data))
	}

	return crashes, nil
}
