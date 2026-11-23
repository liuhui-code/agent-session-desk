// Name cleanup adapted from Agent Deck, Copyright (c) 2025 Ashesh Goplani.
// SPDX-License-Identifier: MIT
// Source: internal/session/groups.go at 80494e527393b7ead00da30018e6ee7e60e9d7a0.
// Modified: stable IDs, explicit bounded validation; no filesystem paths as IDs.
package session

import (
	"fmt"
	"strings"
	"unicode"
	"unicode/utf8"
)

const DefaultGroupID = "default"
const MaxGroups = 256 // Defensive per-scope bound, not a resource budget.
const MaxGroupDepth = 16

type Group struct {
	ID         string `json:"group_id"`
	ParentID   string `json:"parent_id,omitempty"`
	Name       string `json:"name"`
	DefaultDir string `json:"default_dir,omitempty"`
	Revision   int64  `json:"revision"`
	Order      int64  `json:"order"`
}

type Workspace struct {
	Sessions []Session `json:"sessions"`
	Groups   []Group   `json:"groups"`
}

// GroupOrder captures the sibling positions the caller actually observed.
// Names and positions are not terminal or Session/Run addresses.
type GroupOrder struct {
	TargetID string
	Siblings []Group
}

// GroupRemoval is a read-only review snapshot, never a terminal-control token.
type GroupRemoval struct {
	Root, Destination Group
	Groups            []Group
	MemberCount       int64
	Stamp             string
}

func GroupOf(v Session) string {
	if v.GroupID == "" {
		return DefaultGroupID
	}
	return v.GroupID
}

func NormalizeGroupName(value string) (string, error) {
	if len(value) > 4096 || !utf8.ValidString(value) || strings.TrimSpace(value) == "" {
		return "", fmt.Errorf("group name input exceeds bounds, is empty or is not UTF-8")
	}
	for _, r := range value {
		if unicode.IsControl(r) || r == '\u2028' || r == '\u2029' {
			return "", fmt.Errorf("group name must be one line")
		}
	}
	var b strings.Builder
	for _, r := range value {
		if unicode.IsLetter(r) || unicode.IsDigit(r) || r == ' ' || r == '-' || r == '_' {
			b.WriteRune(r)
		} else if r == '/' || r == '\\' || r == '.' {
			b.WriteRune('-')
		}
	}
	name := b.String()
	for strings.Contains(name, "--") {
		name = strings.ReplaceAll(name, "--", "-")
	}
	name = strings.Trim(name, "- ")
	if name == "" {
		name = "unnamed"
	}
	if utf8.RuneCountInString(name) > 50 || len(name) > 200 {
		return "", fmt.Errorf("group name exceeds 50 code points or 200 UTF-8 bytes")
	}
	return name, nil
}

func NormalizeGroupDir(value string) (string, error) {
	if len(value) > 4096 || !utf8.ValidString(value) {
		return "", fmt.Errorf("group directory exceeds bounds or is not UTF-8")
	}
	for _, r := range value {
		if unicode.IsControl(r) || r == '\u2028' || r == '\u2029' {
			return "", fmt.Errorf("group directory must be one line")
		}
	}
	return strings.TrimSpace(value), nil
}
