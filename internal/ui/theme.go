package ui

func effectiveUITheme(value string) string {
	if value == "light" {
		return "light"
	}
	return "dark"
}

func (m *Model) applyUITheme(value string) {
	m.tokens = m.tokens.withTheme(effectiveUITheme(value))
	m.styles = newStyles(m.tokens)
}
