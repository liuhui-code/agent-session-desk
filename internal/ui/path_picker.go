// Adapted from Agent Deck, Copyright (c) 2025 Ashesh Goplani.
// SPDX-License-Identifier: MIT
// Source: pathfuzzy.go, newdialog.go and group_dialog.go at
// 80494e527393b7ead00da30018e6ee7e60e9d7a0. Modified: bounded asynchronous
// work and independent modal ownership; see ADR-0037 and third_party/agent-deck.
package ui

import (
	"context"

	tea "charm.land/bubbletea/v2"
	"github.com/liuhui-code/agent-session-desk/internal/pathsuggest"
)

type pathPicker struct {
	navigated                            bool
	scanned                              bool
	disk                                 []string
	ready                                bool
	deferred                             []tea.Msg
	deferredBytes                        int
	launch, loaded, hidden, active, soft bool
	generation                           uint64
	query                                string
	corpus, matches, cycle               []string
	selected, cycleIndex                 int
	exact, tabPending                    bool
}
type pathRequest struct {
	complete   bool
	owner      *pathPicker
	generation uint64
	query      string
	cancel     context.CancelFunc
}
type pathResult struct {
	loaded     bool
	request    *pathRequest
	recent     []string
	completion pathsuggest.Completion
	err        error
}
type directoryState struct {
	load             func(context.Context) ([]string, error)
	save             func(context.Context, []string) error
	scan             func(context.Context, string) pathsuggest.Completion
	running, pending *pathRequest
	writing          bool
	wanted           []string
}

func (m *Model) newPathPicker(launch bool) *pathPicker {
	p := &pathPicker{launch: launch, cycleIndex: -1}
	for _, g := range m.sidebarGroups() {
		if pathsuggest.Safe(g.DefaultDir) {
			p.corpus = append(p.corpus, g.DefaultDir)
		}
	}
	p.matches = pathsuggest.Filter(p.corpus, "", nil)
	return p
}
func (m *Model) focusedPathPicker() *pathPicker {
	if m.launch != nil && m.launch.focus == 2 && !m.launch.busy {
		return m.launch.paths
	}
	if d := m.groupDialog; d != nil && !d.move && !d.rename && d.focus == 1 && !d.busy && !d.submitted {
		return d.paths
	}
	return nil
}
func (m *Model) ownsPathPicker(p *pathPicker) bool {
	return p != nil && ((m.launch != nil && m.launch.paths == p) || (m.groupDialog != nil && m.groupDialog.paths == p))
}
func (m *Model) requestPaths(p *pathPicker, query string, explicit ...bool) tea.Cmd {
	if p == nil {
		return nil
	}
	p.generation++
	p.query = query
	p.matches = pathsuggest.Filter(p.corpus, query, nil)
	p.selected = 0
	p.navigated = false
	p.exact, p.ready = false, false
	p.scanned = false
	p.disk = nil
	state := &m.workspace.directories
	state.pending = &pathRequest{owner: p, generation: p.generation, query: query, complete: len(explicit) > 0 && explicit[0]}
	if state.running != nil {
		state.running.cancel()
		return nil
	}
	return m.nextPathRequest()
}
func (m *Model) nextPathRequest() tea.Cmd {
	s := &m.workspace.directories
	r := s.pending
	s.pending = nil
	if r == nil || m.ctx.Err() != nil || !m.ownsPathPicker(r.owner) || m.focusedPathPicker() != r.owner {
		return nil
	}
	ctx, cancel := context.WithCancel(m.ctx)
	r.cancel = cancel
	s.running = r
	load, scan, loaded := s.load, s.scan, r.owner.loaded
	if scan == nil {
		scan = pathsuggest.Complete
	}
	return func() tea.Msg {
		result := pathResult{request: r}
		if !loaded && load != nil {
			result.recent, result.err = load(ctx)
		}
		result.loaded = ctx.Err() == nil
		if ctx.Err() == nil && (r.complete || pathsuggest.PathQuery(r.query)) {
			result.completion = scan(ctx, r.query)
		}
		cancel()
		return result
	}
}
func (m *Model) pathResultEvent(msg tea.Msg) (tea.Cmd, bool) {
	r, ok := msg.(pathResult)
	if !ok {
		return nil, false
	}
	s := &m.workspace.directories
	if s.running != r.request {
		return nil, true
	}
	s.running = nil
	p := r.request.owner
	if m.ownsPathPicker(p) {
		if !p.loaded && r.loaded {
			p.loaded = true
			if r.err == nil {
				p.corpus = append(r.recent, p.corpus...)
			} else {
				m.status = "最近目录未载入，仍可编辑路径"
			}
		}
		if p.generation == r.request.generation && m.focusedPathPicker() == p {
			chosen := ""
			index := p.selected
			if p.launch {
				index--
			}
			if p.navigated && index >= 0 && index < len(p.matches) {
				chosen = p.matches[index]
			}
			p.matches = pathsuggest.Filter(p.corpus, p.query, r.completion.Paths)
			if chosen != "" {
				for i, path := range p.matches {
					if path == chosen {
						p.selected = i
						if p.launch {
							p.selected++
						}
						break
					}
				}
			}
			p.exact, p.ready = r.completion.Exact, true
			p.disk = r.completion.Paths
			p.scanned = r.request.complete || pathsuggest.PathQuery(r.request.query)
			last := len(p.matches) - 1
			if p.launch {
				last++
			}
			p.selected = max(0, min(p.selected, last))
			if p.tabPending {
				p.tabPending = false
				cmd := m.completePathTab(p, r.completion.Paths)
				cmds := []tea.Cmd{cmd, m.nextPathRequest()}
				queued := p.deferred
				p.deferred = nil
				p.deferredBytes = 0
				for _, event := range queued {
					if m.launch == nil || m.launch.paths != p {
						break
					}
					c, _ := m.launchEvent(event)
					cmds = append(cmds, c)
				}
				return tea.Batch(cmds...), true
			}
		}
	}
	return m.nextPathRequest(), true
}

// Passive receipts must release the single worker even when another modal owns input.
func (m *Model) passiveEvent(msg tea.Msg) (tea.Cmd, bool) {
	if m.colorProfileEvent(msg) {
		return nil, true
	}
	if cmd, ok := m.systemThemeEvent(msg); ok {
		return cmd, true
	}
	if m.workspace != nil {
		if cmd, ok := m.pathResultEvent(msg); ok {
			return cmd, true
		}
		if cmd, ok := m.recentDirectoryEvent(msg); ok {
			return cmd, true
		}
	}
	return nil, false
}
func (m *Model) pathEdited(p *pathPicker, query string) tea.Cmd {
	p.hidden, p.active, p.soft = false, false, false
	p.cycle = nil
	p.cycleIndex = -1
	p.tabPending = false
	p.deferred = nil
	p.deferredBytes = 0
	return m.requestPaths(p, query)
}
func (m *Model) pathValue(p *pathPicker) string {
	if p.launch {
		return m.launch.fields[2].Value()
	}
	return m.groupDialog.inputs[1].Value()
}
func (m *Model) setPathValue(p *pathPicker, value string) {
	if p.launch {
		m.launch.fields[2].SetValue(value)
		m.launch.fields[2].CursorEnd()
	} else {
		m.groupDialog.inputs[1].SetValue(value)
		m.groupDialog.inputs[1].CursorEnd()
	}
}
func (m *Model) focusPathInput(p *pathPicker) tea.Cmd {
	if p.launch {
		return m.launch.fields[2].Focus()
	}
	return m.groupDialog.inputs[1].Focus()
}

func (m *Model) cancelPathPicker(p *pathPicker) {
	if p == nil {
		return
	}
	s := &m.workspace.directories
	p.tabPending = false
	p.deferred = nil
	p.deferredBytes = 0
	if s.pending != nil && s.pending.owner == p {
		s.pending = nil
	}
	if s.running != nil && s.running.owner == p {
		s.running.cancel()
	}
}
func (m *Model) reconcilePathViewport() tea.Cmd {
	p := m.focusedPathPicker()
	if p == nil || m.pathPopupAvailable() {
		return nil
	}
	p.active, p.soft = false, false
	return m.focusPathInput(p)
}
