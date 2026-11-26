package ui

import (
	"context"
	"fmt"

	tea "charm.land/bubbletea/v2"
	"github.com/liuhui-code/agent-session-desk/internal/terminal"
)

const workspaceInputCapacity = 64 // defensive input bound, not an agent queue

type workspaceInput struct {
	deliver          func(context.Context, terminal.Connection) error
	kind, text       string
	button           string
	modifiers        uint8
	width, height    int
	x, y             int
	anchorX, anchorY int
	up               bool
	reply            chan error
	hover            *workspaceProgramPointer
}
type workspaceInputDone struct {
	generation uint64
	err        error
}

func (m *Model) workspaceWrite(action workspaceInput) tea.Cmd {
	w := m.workspace
	if len(action.text) > terminal.MaxInputBytes {
		m.status = "Terminal input exceeds byte limit; not sent"
		return nil
	}
	if (w.busy || w.transition == workspaceSwitch) && (action.kind == "text" || action.kind == "key" || action.kind == "paste") {
		if len(w.buffered) >= workspaceInputCapacity {
			m.status = "Connecting input buffer full; input not sent"
			return nil
		}
		w.buffered = append(w.buffered, action)
		return nil
	}
	if w.conn == nil || w.inputs == nil || w.err != nil || w.busy {
		m.status = "Terminal disconnected; input not sent"
		return nil
	}
	action.reply = make(chan error, 1)
	if w.programPointer != nil && len(w.inputs) >= cap(w.inputs)-1 {
		m.status = "Mouse input queue full; release position reserved for the original target"
		return nil
	}
	select {
	case w.inputs <- action:
		w.pending++
	default:
		m.status = "Terminal input queue full; input not sent"
		return nil
	}
	generation, ctx := w.generation, w.ctx
	return func() tea.Msg {
		select {
		case err := <-action.reply:
			return workspacePointerReceipt(generation, err, action.hover)
		case <-ctx.Done():
			return workspacePointerReceipt(generation, ctx.Err(), action.hover)
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
				case "delivery":
					err = action.deliver(ctx, conn)
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
				case "program-mouse":
					if native, ok := conn.(terminal.ProgramMouseConnection); ok {
						err = native.Mouse(ctx, action.text, action.button, action.modifiers, action.x, action.y)
					} else if native, ok := conn.(terminal.PointerConnection); ok && action.button == "left" && action.modifiers == 0 {
						err = native.Pointer(ctx, action.text, action.x, action.y)
					} else {
						err = fmt.Errorf("Program mouse event unsupported by this connection")
					}
				case "pointer":
					if native, ok := conn.(terminal.PointerConnection); ok {
						err = native.Pointer(ctx, action.text, action.x, action.y)
					} else {
						err = fmt.Errorf("Native mouse input unsupported by this connection")
					}
				case "copy-drag":
					if native, ok := conn.(terminal.CopyDragConnection); ok {
						err = native.BeginCopyDrag(ctx, action.anchorX, action.anchorY, action.x, action.y)
					} else {
						err = fmt.Errorf("Live terminal selection unsupported by this connection")
					}
				case "copy-mode":
					if native, ok := conn.(terminal.CopyModeConnection); ok {
						err = native.EnterCopyMode(ctx)
					} else {
						err = fmt.Errorf("Native selection/search unsupported by this connection")
					}
				default:
					err = fmt.Errorf("Unknown terminal input type")
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
		m.status = "Key combination unsupported; not sent"
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
