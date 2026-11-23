// Package appearance observes a UI theme without changing terminal or OS state.
// Terminal declaration intent adapted from Agent Deck userconfig.go at
// 80494e527393b7ead00da30018e6ee7e60e9d7a0, Copyright (c) 2025 Ashesh
// Goplani, MIT. Modified: strict bounded parsing and sustained declaration
// priority. See third_party/agent-deck/LICENSE and THIRD_PARTY_NOTICES.md.
package appearance

import (
	"context"
	"os"
	"strconv"
	"strings"
)

type Observation struct {
	Theme, Source string
	Err           error
}
type Connection interface {
	Next(context.Context) (Observation, error)
	Close() error
}

// Open binds all source resources to ctx, including when no caller receives the
// result. Close is idempotent and joins teardown; Next permits one reader.
func Open(ctx context.Context) (Connection, error) {
	if e := ctx.Err(); e != nil {
		return nil, e
	}
	if theme, ok := terminalTheme(os.Getenv("COLORFGBG")); ok {
		lifetime, cancel := context.WithCancel(ctx)
		s := &stream{events: make(chan Observation, 1), done: make(chan struct{}), stop: cancel}
		context.AfterFunc(lifetime, func() { close(s.done) })
		s.publish(Observation{Theme: theme, Source: "terminal"})
		return s, nil
	}
	return openSystem(ctx)
}

func terminalTheme(value string) (string, bool) {
	i := strings.LastIndexByte(value, ';')
	if i < 0 {
		return "", false
	}
	bg := value[i+1:]
	if bg == "" {
		return "", false
	}
	for _, c := range bg {
		if c < '0' || c > '9' {
			return "", false
		}
	}
	v, e := strconv.Atoi(bg)
	if e != nil || v > 15 {
		return "", false
	}
	if v < 8 {
		return "dark", true
	}
	return "light", true
}
