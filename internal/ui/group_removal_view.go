// Layout adapted from Agent Deck, Copyright (c) 2025 Ashesh Goplani.
// SPDX-License-Identifier: MIT
// Source: internal/ui/confirm_dialog.go at 80494e527393b7ead00da30018e6ee7e60e9d7a0.
// Modified: v2 total widths, localized reviewed counts, token styles and bounded
// viewports; no terminal or Session/Run deletion action.
package ui

import (
	"fmt"
	"strings"

	"charm.land/lipgloss/v2"
	"github.com/liuhui-code/agent-session-desk/internal/session"
)

func (m *Model) groupRemovalWidths() (outer, content int) {
	t := m.tokens.confirm
	outer = t.preferredWidth
	if m.width < t.responsiveBelow {
		outer = m.width - t.narrowGutter
	}
	outer = max(1, min(m.width, outer))
	content = max(0, outer-m.styles.confirmBox.GetHorizontalFrameSize())
	return
}

func (m *Model) groupRemovalBox() (string, bool) {
	d, t := m.groupRemoval, m.tokens.confirm
	outer, width := m.groupRemovalWidths()
	name := d.target.Name
	if d.plan.Root.ID != "" {
		name = d.plan.Root.Name
	}
	title := t.titlePrefix + "删除分组？"
	warning := "将删除分组：\n\n" + strings.Repeat(" ", t.nameIndent) + t.nameOpen + label(name) + t.nameClose
	details := "正在核对分组与成员…"
	if !d.loading && d.plan.Stamp == "" && d.err != nil {
		details = "无法核对影响范围，请返回刷新。"
	}
	if d.plan.Stamp != "" {
		details = fmt.Sprintf("%s将移除 %d 个分组（含子组）\n%s%d 个会话收纳到%s%s%s\n%s终端不停止，输出与历史保留", t.bullet, len(d.plan.Groups), t.bullet, d.plan.MemberCount, t.nameOpen, label(d.plan.Destination.Name), t.nameClose, t.bullet)
	}
	button := func(text string, focused, destructive bool) string {
		style := m.styles.confirmCancel
		prefix := t.idlePrefix
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
	buttons := lipgloss.JoinHorizontal(lipgloss.Top, button("删除", d.focus == 0, true), strings.Repeat(" ", t.buttonGap), button("取消", d.focus == 1, false))
	if lipgloss.Width(buttons) > width {
		buttons = lipgloss.JoinVertical(lipgloss.Left, button("删除", d.focus == 0, true), button("取消", d.focus == 1, false))
	}
	hint := "y 删除 · n 取消 · ←/→ 选择 · Enter · Esc"
	if d.loading {
		hint = "正在核对影响范围 · Esc 取消"
	} else if d.busy {
		hint = "正在收纳成员并移除结构…"
	} else if d.submitted {
		hint = "Esc 返回并刷新核对"
	}
	titleStyle, boxStyle := m.styles.confirmTitle, m.styles.confirmBox
	if d.target.ID == session.DefaultGroupID {
		title = t.titlePrefix + "无法删除默认分组"
		warning = t.nameOpen + label(name) + t.nameClose + " 是默认分组，不能删除。\n\n会话需要保留归属。"
		details = ""
		buttons = button("确定", true, false)
		hint = "Enter / Esc / o 返回"
		titleStyle, boxStyle = m.styles.confirmNoticeTitle, m.styles.confirmNoticeBox
	}
	errLine := ""
	if d.err != nil {
		errLine = m.styles.confirmError.Width(width).Render(label(d.err.Error()))
	}
	content := lipgloss.JoinVertical(lipgloss.Left, titleStyle.Width(width).Render(title), m.styles.confirmWarning.Width(width).Render(warning), m.styles.confirmDetails.Width(width).Render(details), errLine, buttons, m.styles.confirmHint.Width(width).Render(hint))
	box := boxStyle.Width(outer).Render(content)
	bw, bh := lipgloss.Size(box)
	return box, bw <= m.width && bh <= m.height && width > 0
}

func (m *Model) renderGroupRemoval() string {
	if m.width <= 0 || m.height <= 0 {
		return ""
	}
	box, fits := m.groupRemovalBox()
	if !fits {
		outer, width := m.groupRemovalWidths()
		compact := m.styles.confirmDetails.Render("请放大窗口核对完整提示") + "\n" + m.styles.confirmHint.Render("Esc 取消")
		box = m.styles.confirmBox.Width(outer).Render(fitFrame(compact, width, max(0, m.height-m.styles.confirmBox.GetVerticalFrameSize())))
	}
	bw, bh := lipgloss.Size(box)
	return overlayLaunchDialog(fitFrame("", m.width, m.height), box, max(0, (m.width-bw)/2), max(0, (m.height-bh)/2), m.width, m.height)
}
