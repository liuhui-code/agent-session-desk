package ui

func (m *Model) systemThemeCaption() string {
	if m.densitySettings == nil || uiThemes[m.densitySettings.themeChoice] != "system" {
		return ""
	}
	s := &m.systemTheme
	if s.choice != "system" {
		return "保存成功后跟随系统"
	}
	if s.blocked {
		return "主题观察器未能关闭；重启界面后重试"
	}
	current := "Dark"
	if s.resolved == "light" {
		current = "Light"
	}
	if s.known {
		if s.source == "terminal" {
			return "跟随终端声明 · 当前 " + current
		}
		return "跟随 Windows 应用主题 · 当前 " + current
	}
	if s.resolved == "" {
		return "系统主题未知 · 暂用 Dark · r 重试"
	}
	return "系统主题未知 · 保留 " + current + " · r 重试"
}
