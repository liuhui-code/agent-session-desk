package ui

import (
	"fmt"
	"strings"
	"time"

	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"
)

func (m *Model) renderHistory(_ int) string {
	h := &m.history
	title := m.renderPanelTitle(fmt.Sprintf("历史记录 · 最近 %d 条", historyLimit), m.width)
	if h.busy {
		return title + "\n" + m.styles.dim.Render("正在读取历史…")
	}
	if h.err != nil {
		return title + "\n" + m.styles.error.Render("读取历史失败："+label(h.err.Error()))
	}
	if len(h.rows) == 0 {
		return title + "\n" + m.styles.dim.Render("暂无历史记录")
	}
	if h.detail {
		lines := m.historyDetailLines()
		offset := min(h.offset, max(0, len(lines)-m.historyPageSize()))
		return title + "\n" + m.styles.text.Render(strings.Join(lines[offset:min(len(lines), offset+m.historyPageSize())], "\n"))
	}
	lines := []string{}
	size := m.historyPageSize()
	start := (h.cursor / size) * size
	for i := start; i < min(len(h.rows), start+size); i++ {
		r := h.rows[i]
		prefix := strings.Repeat(" ", m.tokens.layout.rowPrefixWidth)
		if i == h.cursor {
			prefix = m.tokens.marks.selected + strings.Repeat(" ", max(0, m.tokens.layout.rowPrefixWidth-lipgloss.Width(m.tokens.marks.selected)))
		}
		row := cellTruncate(prefix+label(r.Session.Title), m.width, m.tokens.marks.ellipsis)
		if i == h.cursor {
			row = m.styles.highlight.Render(row)
		} else {
			row = m.styles.text.Render(row)
		}
		lines = append(lines, row)
	}
	return title + "\n" + strings.Join(lines, "\n")
}

// Each metadata value becomes inert, bounded text before display-cell wrapping.
func (m *Model) historyDetailLines() []string {
	h := &m.history
	if len(h.rows) == 0 {
		return nil
	}
	r := h.rows[h.cursor]
	fields := []string{"只读记录 · 当前在线状态未知", "名称：" + label(r.Session.Title), "Session：" + label(r.Session.ID), "Run：" + label(r.Session.RunID), "Backend：" + label(r.Session.BackendID), "程序：" + label(r.Session.Command), "首次记录：" + r.RecordedAt.UTC().Format(time.RFC3339Nano), "agent 结果：未知"}
	stop := "未记录（不代表未停止）"
	if r.StopConfirmedAt != nil {
		stop = r.StopConfirmedAt.UTC().Format(time.RFC3339Nano)
	}
	fields = append(fields, "停止确认："+stop)
	if r.PaneExit == nil {
		fields = append(fields, "pane 退出观测：未记录")
	} else {
		fields = append(fields, "pane："+label(r.PaneExit.PaneID), "观测来源："+label(r.PaneExit.Source), "退出观测时间："+r.PaneExit.ObservedAt.UTC().Format(time.RFC3339Nano))
	}
	fields = append(fields, "退出码：未知")
	return strings.Split(ansi.Hardwrap(strings.Join(fields, "\n"), max(1, m.width), true), "\n")
}

func (m *Model) renderHistoryHelpBar() string {
	action := "选择"
	if m.history.detail {
		action = "滚动"
	}
	parts := []string{m.styles.menuKey("↑↓/PgUp/PgDown", action)}
	if !m.history.detail {
		parts = append(parts, m.styles.menuKey("Enter", "详情"))
	}
	parts = append(parts, m.styles.menuKey("r", "刷新"), m.styles.menuKey("Esc/h", "返回"), m.styles.menuKey("q", "退出"))
	return strings.Join(parts, m.tokens.itemGap())
}
