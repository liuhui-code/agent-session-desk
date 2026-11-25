package ui

type helpSection struct {
	title string
	items [][2]string
}

func workspaceHelpSections() []helpSection {
	return []helpSection{
		{"Tabs and terminals", [][2]string{
			{"F6", "Toggle navigation/terminal focus"},
			{"Ctrl+PgUp / Ctrl+PgDn", "Switch tabs; preserve the terminal draft"},
			{"Enter", "Open the selected session's original run"},
			{"n / + button", "Create an agent, shell or CLI session"},
			{"x / × button", "Close only the view; keep the run alive"},
			{"d (navigation focus)", "Confirm stopping the selected session; preserve tab and history"},
			{"[ / ]", "Move the active tab left or right"},
			{"Paste", "Use the host terminal paste shortcut"},
			{"q / Ctrl+C", "Sent to the target program with terminal focus"},
		}},
		{"Selection and search (navigation focus)", [][2]string{
			{"↑ / ↓ / k / j", "Select a list item; click Select, double-click Open"},
			{"Home / End", "Select first/last item"},
			{"/", "Search local session names or programs"},
			{"r", "Rename selected session/group; retry when disconnected"},
			{"Ctrl+R", "Refresh backend observations"},
		}},
		{"Groups (navigation focus)", [][2]string{
			{"Alt+V", "Toggle grouped/flat navigation"},
			{"1–9", "Select root group"},
			{"Enter / → / Tab", "Expand/collapse group; ← Return to group and collapse"},
			{"g", "Create a group or subgroup with an optional default directory"},
			{"M", "Move the selected session to a group"},
			{"Shift/Ctrl+↑↓ / K / J / + / -", "Reorder sessions within a group or sibling groups, then save"},
			{"d", "Review group removal; move members to the default group. Cancel by default"},
		}},
		{"Interface, history and auxiliary actions", [][2]string{
			{"S", "Theme and density; apply after saving"},
			{"< / >", "Adjust and save split; drag saves on release"},
			{"=", "Restore and save automatic split"},
			{"h", "Read-only history; no terminal input"},
			{"v", "Native selection/search; Space Select, Enter Copy, / Search"},
			{"V", "Copy the selected session's current screen, excluding history"},
			{"Left drag", "Drag to select when the program has not requested mouse input"},
			{"Left button", "Mouse-enabled programs receive press, drag and release"},
			{"Wheel", "Scroll the list or native terminal history under the pointer"},
			{"a", "Auxiliary full-screen attach; Ctrl+b then d to return"},
			{"?", "Open quick start; Tab shows all shortcuts"},
			{"q / Ctrl+C", "Quit from navigation; keep background runs alive"},
		}},
	}
}
