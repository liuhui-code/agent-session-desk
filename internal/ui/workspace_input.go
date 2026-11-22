package ui

import (
	"context"
	"fmt"

	tea "charm.land/bubbletea/v2"
	"github.com/liuhui-code/agent-session-desk/internal/terminal"
)

const workspaceInputCapacity = 64 // defensive input bound, not an agent queue

type workspaceInput struct {
	kind, text    string
	width, height int
	x, y          int
	up            bool
	reply         chan error
}
type workspaceInputDone struct {
	generation uint64
	err        error
}

func (m *Model) workspaceWrite(action workspaceInput) tea.Cmd {
	w := m.workspace
	if len(action.text) > terminal.MaxInputBytes {
		m.status = "终端输入超过字节上限；此输入未发送"
		return nil
	}
	if (w.busy || w.transition == workspaceSwitch) && (action.kind == "text" || action.kind == "key" || action.kind == "paste") {
		if len(w.buffered) >= workspaceInputCapacity {
			m.status = "连接中输入已达上限；此输入未发送"
			return nil
		}
		w.buffered = append(w.buffered, action)
		return nil
	}
	if w.conn == nil || w.inputs == nil || w.err != nil || w.busy {
		m.status = "终端尚未连接；此输入未发送"
		return nil
	}
	action.reply = make(chan error, 1)
	select {
	case w.inputs <- action:
		w.pending++
	default:
		m.status = "终端输入队列已满；此输入未发送"
		return nil
	}
	generation, ctx := w.generation, w.ctx
	return func() tea.Msg {
		select {
		case err := <-action.reply:
			return workspaceInputDone{generation, err}
		case <-ctx.Done():
			return workspaceInputDone{generation, ctx.Err()}
		}
	}
}

func startWorkspaceWriter(ctx context.Context, conn terminal.Connection, inputs <-chan workspaceInput) {
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case action := <-inputs:
				if ctx.Err() != nil {
					return
				}
				var err error
				switch action.kind {
				case "text":
					err = conn.SendText(ctx, action.text)
				case "key":
					err = conn.SendKey(ctx, action.text)
				case "paste":
					err = conn.Paste(ctx, action.text)
				case "resize":
					err = conn.Resize(ctx, action.width, action.height)
				case "refresh":
					err = conn.Refresh(ctx)
				case "scroll":
					err = conn.Scroll(ctx, action.x, action.y, action.up)
				default:
					err = fmt.Errorf("未知终端输入类型")
				}
				action.reply <- err
				if err != nil {
					return
				}
			}
		}
	}()
}

func (m *Model) workspaceTerminalKey(msg tea.KeyPressMsg) tea.Cmd {
	key := msg.Key()
	if key.Text != "" && key.Mod&(tea.ModCtrl|tea.ModAlt) == 0 {
		return m.workspaceWrite(workspaceInput{kind: "text", text: key.Text})
	}
	names := map[rune]string{
		tea.KeyEnter: "Enter", tea.KeyTab: "Tab", tea.KeyBackspace: "BSpace", tea.KeyEscape: "Escape",
		tea.KeyLeft: "Left", tea.KeyRight: "Right", tea.KeyUp: "Up", tea.KeyDown: "Down",
		tea.KeyHome: "Home", tea.KeyEnd: "End", tea.KeyPgUp: "PPage", tea.KeyPgDown: "NPage",
		tea.KeyDelete: "DC", tea.KeyInsert: "IC",
	}
	name := names[key.Code]
	if key.Code >= tea.KeyF1 && key.Code <= tea.KeyF12 {
		name = fmt.Sprintf("F%d", key.Code-tea.KeyF1+1)
	}
	if name == "" && key.Code >= 'a' && key.Code <= 'z' {
		name = string(key.Code)
	}
	if name == "" && key.Code >= '0' && key.Code <= '9' {
		name = string(key.Code)
	}
	if name == "" {
		m.status = "当前组合键尚未支持；未发送"
		return nil
	}
	prefix := ""
	if key.Mod&tea.ModCtrl != 0 {
		prefix += "C-"
	}
	if key.Mod&tea.ModAlt != 0 {
		prefix += "M-"
	}
	if key.Mod&tea.ModShift != 0 {
		prefix += "S-"
	}
	return m.workspaceWrite(workspaceInput{kind: "key", text: prefix + name})
}
