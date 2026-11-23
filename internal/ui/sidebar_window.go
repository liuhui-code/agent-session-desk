package ui

// A group header occupies one row while cards occupy density-dependent rows.
// Rendering and mouse hit-testing consume these same visible rectangles.
type sidebarHit struct{ index, top, height int }
type sidebarWindow struct {
	start, end, leading int
	rows                []sidebarHit
	remaining           int
}

func (m *Model) sidebarNavigationWindow(width, height int) sidebarWindow {
	if height <= 0 {
		return sidebarWindow{}
	}
	card := m.workspaceSidebarRows(width, height)
	if !m.groupedSidebar() {
		old := m.sessionListWindowForRows(height, card)
		window := sidebarWindow{start: old.start, end: old.end, leading: old.leading, remaining: len(m.sessions) - old.end}
		top := old.leading
		for i := old.start; i < old.end && top < height; i++ {
			drawn := min(card, height-top)
			window.rows = append(window.rows, sidebarHit{i, top, drawn})
			top += drawn
		}
		return window
	}
	count := 1
	if !m.sidebarCollapsed {
		count += len(m.sessions)
	}
	itemHeight := func(i int) int {
		if i == 0 {
			return 1
		}
		return card
	}
	selected := 0
	if !m.sidebarGroupSelected() {
		selected = min(count-1, m.cursor+1)
	}
	build := func(start int) sidebarWindow {
		window := sidebarWindow{start: start, end: start}
		top := 0
		if start > 0 && height >= m.tokens.layout.aboveMinHeight {
			window.leading = 1
			top = 1
		}
		for i := start; i < count && top < height; i++ {
			drawn := itemHeight(i)
			if top+drawn > height {
				if len(window.rows) > 0 {
					break
				}
				drawn = height - top
			}
			window.rows = append(window.rows, sidebarHit{i - 1, top, drawn})
			top += drawn
			window.end = i + 1
		}
		window.remaining = count - window.end
		return window
	}
	start := max(0, min(m.viewOffset, selected))
	window := build(start)
	if selected >= window.end {
		start = selected
		budget := height
		if start > 0 && height >= m.tokens.layout.aboveMinHeight {
			budget--
		}
		used := min(budget, itemHeight(selected))
		for start > 0 && used+itemHeight(start-1) <= budget {
			start--
			used += itemHeight(start)
		}
		window = build(start)
	}
	return window
}
