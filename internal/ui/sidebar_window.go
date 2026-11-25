package ui

// A group header occupies one row while cards occupy density-dependent rows.
// Rendering and mouse hit-testing consume these same visible rectangles.
type sidebarHit struct {
	index, top, height, level int
	groupID                   string
}
type sidebarWindow struct {
	start, end, leading int
	rows                []sidebarHit
	remaining           int
}

func (m *Model) sidebarNavigationWindow(width, height int) sidebarWindow {
	if height <= 0 {
		return sidebarWindow{}
	}
	card := m.workspaceSidebarRows(width, height) + m.tokens.sidebar.cardGap
	if !m.groupedSidebar() {
		old := m.sessionListWindowForRows(height, card)
		window := sidebarWindow{start: old.start, end: old.end, leading: old.leading, remaining: len(m.sessions) - old.end}
		top := old.leading
		for i := old.start; i < old.end && top < height; i++ {
			drawn := min(card, height-top)
			window.rows = append(window.rows, sidebarHit{index: i, top: top, height: drawn})
			top += drawn
		}
		return window
	}
	items := m.sidebarItems()
	count := len(items)
	if count == 0 {
		return sidebarWindow{}
	}
	itemHeight := func(i int) int {
		if items[i].index < 0 {
			return min(height, m.tokens.sidebar.groupRows)
		}
		return card
	}
	selected := m.sidebarSelectedItem(items)
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
			item := items[i]
			window.rows = append(window.rows, sidebarHit{index: item.index, top: top, height: drawn, level: item.level, groupID: item.groupID})
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
