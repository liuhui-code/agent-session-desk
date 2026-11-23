package ui

import (
	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/colorprofile"
)

// Capability notifications are display-only, including while a modal is open.
// Unknown/invalid receipts do not discard the last known capability.
func (m *Model) colorProfileEvent(msg tea.Msg) bool {
	v, ok := msg.(tea.ColorProfileMsg)
	if !ok {
		return false
	}
	if v.Profile <= colorprofile.Unknown || v.Profile > colorprofile.TrueColor || m.colorProfile == v.Profile {
		return true
	}
	m.colorProfile = v.Profile
	m.restyleUI()
	return true
}

func (m *Model) applyViewColors(view *tea.View) {
	if m.colorProfile == colorprofile.TrueColor {
		view.BackgroundColor, view.ForegroundColor = m.tokens.colors.bg, m.tokens.colors.text
	}
}
