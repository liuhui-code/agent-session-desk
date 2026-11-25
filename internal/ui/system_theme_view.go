package ui

func (m *Model) systemThemeCaption() string {
	if m.densitySettings == nil || uiThemes[m.densitySettings.themeChoice] != "system" {
		return ""
	}
	s := &m.systemTheme
	if s.choice != "system" {
		return "Follow the system after saving"
	}
	if s.blocked {
		return "Theme observer did not close; restart the UI to retry"
	}
	current := "Dark"
	if s.resolved == "light" {
		current = "Light"
	}
	if s.known {
		if s.source == "terminal" {
			return "Terminal theme · Current " + current
		}
		return "Windows app theme · Current " + current
	}
	if s.resolved == "" {
		return "System theme unknown · Using Dark · r Retry"
	}
	return "System theme unknown · Keeping " + current + " · r retry"
}
