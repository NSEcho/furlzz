package mutator

func (m *Mutator) getFuzzedInput() string {
	if m.multipleRounds {
		if m.lastInput == "" {
			m.lastInput = m.fetchInput()
		}
		return m.lastInput
	}
	return m.fetchInput()
}

func (m *Mutator) fetchInput() string {
	if m.fuzzIdx == -1 || len(m.validInputs) == 0 {
		return m.input
	}
	k := m.r.Intn(len(m.validInputs))
	return m.validInputs[k]
}
