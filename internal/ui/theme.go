package ui

func effectiveUITheme(value string) string {
	if value == "light" || value == "system" {
		return value
	}
	return "dark"
}

func (m *Model) applyUITheme(value string) {
	s := &m.systemTheme
	value = effectiveUITheme(value)
	if s.choice != value {
		s.choice, s.failed = value, false
		if s.op != nil {
			s.op.invalid = true
			s.op.cancel()
		}
	}
	m.drawUITheme()
}

func (m *Model) drawUITheme() {
	value := m.systemTheme.choice
	if value == "system" {
		value = m.systemTheme.resolved
	}
	if value != "light" {
		value = "dark"
	}
	m.tokens = m.tokens.withTheme(value)
	m.styles = newStyles(m.tokens)
	// SetStyles keeps the widget's value, cursor and focus intact.
	if m.launch != nil {
		for i := range m.launch.fields {
			m.launch.fields[i].SetStyles(m.styles.formInput)
		}
	}
	if m.rename != nil {
		m.rename.input.SetStyles(m.styles.renameInput)
	}
	if m.search != nil {
		m.search.input.SetStyles(searchInputStyles(m.tokens))
	}
	if m.groupDialog != nil {
		for i := range m.groupDialog.inputs {
			m.groupDialog.inputs[i].SetStyles(m.styles.groupInput)
		}
		if m.groupDialog.rename {
			m.groupDialog.inputs[0].SetStyles(m.styles.renameInput)
		}
	}
}
