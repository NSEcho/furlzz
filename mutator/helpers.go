package mutator

func (m *Mutator) getFuzzedInput() []byte {
	if m.multipleRounds {
		if len(m.lastInput) == 0 {
			m.lastInput = m.fetchInput()
		}
		return m.lastInput
	}
	return m.fetchInput()
}

func (m *Mutator) fetchInput() []byte {
	if m.fuzzIdx == -1 || len(m.validInputs) == 0 {
		return m.input
	}
	k := m.r.Intn(len(m.validInputs))
	return m.validInputs[k]
}
