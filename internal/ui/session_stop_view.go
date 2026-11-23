// Confirmation layout adapted from Agent Deck, Copyright (c) 2025 Ashesh
// Goplani. SPDX-License-Identifier: MIT
// Source: confirm_dialog.go at 80494e527393b7ead00da30018e6ee7e60e9d7a0.
// Modified: complete target review, localized stop-only text and existing tokens.
package ui

import (
	"strings"

	"charm.land/lipgloss/v2"
)

func (m *Model) sessionStopBox() (string, bool) {
	d, t := m.sessionStop, m.tokens.confirm
	outer, width := m.groupRemovalWidths()
	title := t.titlePrefix + "停止会话？"
	warning := "将停止终端及其程序：\n\n" + strings.Repeat(" ", t.nameIndent) + t.nameOpen + label(d.target.Title) + t.nameClose
	details := "Session：" + label(d.target.ID) + "\nRun：" + label(d.target.RunID) + "\n程序：" + label(d.target.Command) + "\n" + t.bullet + "未保存的工作可能丢失\n" + t.bullet + "Tab 视图、名称、分组与历史保留"
	button := func(text string, focused, destructive bool) string {
		style, prefix := m.styles.confirmCancel, t.idlePrefix
		if destructive {
			style = m.styles.confirmDelete
		}
		if focused {
			prefix = t.focusedPrefix
			if destructive {
				style = m.styles.confirmDeleteFocused
			} else {
				style = m.styles.confirmCancelFocused
			}
		}
		return style.Render(prefix + text)
	}
	buttons := lipgloss.JoinHorizontal(lipgloss.Top, button("停止", d.focus == 0, true), strings.Repeat(" ", t.buttonGap), button("取消", d.focus == 1, false))
	if lipgloss.Width(buttons) > width {
		buttons = lipgloss.JoinVertical(lipgloss.Left, button("停止", d.focus == 0, true), button("取消", d.focus == 1, false))
	}
	hint := "y 停止 · n 取消 · ←/→ 选择 · Enter · Esc"
	if d.busy {
		hint = "正在核对并停止原 Run…"
	} else if d.submitted {
		hint = "Esc 返回并刷新核对"
	}
	errLine := ""
	if d.err != nil {
		errLine = m.styles.confirmError.Width(width).Render(label(d.err.Error()))
	}
	content := lipgloss.JoinVertical(lipgloss.Left, m.styles.confirmTitle.Width(width).Render(title), m.styles.confirmWarning.Width(width).Render(warning), m.styles.confirmDetails.Width(width).Render(details), errLine, buttons, m.styles.confirmHint.Width(width).Render(hint))
	box := m.styles.confirmBox.Width(outer).Render(content)
	bw, bh := lipgloss.Size(box)
	return box, width > 0 && bw <= m.width && bh <= m.height
}

func (m *Model) renderSessionStop() string {
	if m.width <= 0 || m.height <= 0 {
		return ""
	}
	box, fits := m.sessionStopBox()
	if !fits {
		outer, width := m.groupRemovalWidths()
		hint := "Esc 取消"
		if m.sessionStop.busy {
			hint = "正在停止，请等待结果"
		} else if m.sessionStop.submitted {
			hint = "结果未确认 · Esc 返回刷新核对"
		}
		compact := m.styles.confirmDetails.Render("请放大窗口核对完整提示") + "\n" + m.styles.confirmHint.Render(hint)
		box = m.styles.confirmBox.Width(outer).Render(fitFrame(compact, width, max(0, m.height-m.styles.confirmBox.GetVerticalFrameSize())))
	}
	bw, bh := lipgloss.Size(box)
	return overlayLaunchDialog(fitFrame("", m.width, m.height), box, max(0, (m.width-bw)/2), max(0, (m.height-bh)/2), m.width, m.height)
}
