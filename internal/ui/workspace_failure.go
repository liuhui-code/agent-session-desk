package ui

// Failure ends this view's input authority. Keep only its already observed
// screen for read-only inspection; unsent input is never replayed on retry.
func (m *Model) failWorkspace(err error) {
	if m.sessionOrder != nil {
		m.sessionOrder.quitting = false
	}
	if m.groupOrder != nil {
		m.groupOrder.quitting = false
	}
	w := m.workspace
	w.disconnect()
	w.err = err
	w.focused = false
	w.buffered = nil
	w.transition = workspaceIdle
	m.status = ""
}
