// Portions adapted from Agent Deck, Copyright (c) 2025 Ashesh Goplani.
// SPDX-License-Identifier: MIT
// Source: internal/session/discovery.go at
// 80494e527393b7ead00da30018e6ee7e60e9d7a0: FilterByQuery/localQueryDistance.
// Modified: Session title/program only; no inferred agent-status filters.
// See third_party/agent-deck/LICENSE.
package ui

import (
	"sort"
	"strings"

	"github.com/liuhui-code/agent-session-desk/internal/session"
)

func filterSessionSearch(items []session.Session, query string) []session.Session {
	query = strings.ToLower(strings.TrimSpace(query))
	if query == "" {
		return append([]session.Session(nil), items...)
	}
	type scored struct {
		item     session.Session
		distance int
	}
	var matches []scored
	for _, item := range items {
		best := -1
		for _, field := range []string{item.Title, item.Command} {
			if distance, ok := sessionSearchDistance(label(field), query); ok && (best < 0 || distance < best) {
				best = distance
			}
		}
		if best >= 0 {
			matches = append(matches, scored{item, best})
		}
	}
	sort.SliceStable(matches, func(i, j int) bool { return matches[i].distance < matches[j].distance })
	results := make([]session.Session, len(matches))
	for i, match := range matches {
		results[i] = match.item
	}
	return results
}

func sessionSearchDistance(text, query string) (int, bool) {
	text = strings.ToLower(text)
	if strings.Contains(text, query) {
		return 0, true
	}
	if text == "" {
		return 0, false
	}
	queryRunes := []rune(query)
	tolerance := 2
	if len(queryRunes) <= 3 {
		return 0, false
	}
	if len(queryRunes) <= 5 {
		tolerance = 1
	}
	previous := make([]int, len(queryRunes)+1)
	current := make([]int, len(previous))
	for i := range previous {
		previous[i] = i
	}
	best := tolerance + 1
	for _, char := range text {
		current[0] = 0
		for i, queryChar := range queryRunes {
			cost := 0
			if char != queryChar {
				cost = 1
			}
			current[i+1] = min(previous[i+1]+1, current[i]+1, previous[i]+cost)
		}
		best = min(best, current[len(queryRunes)])
		previous, current = current, previous
	}
	return best, best <= tolerance
}
