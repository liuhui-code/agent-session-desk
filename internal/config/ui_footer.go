// Footer normalization adapted from Agent Deck userconfig.go at
// 80494e527393b7ead00da30018e6ee7e60e9d7a0, Copyright (c) 2025 Ashesh
// Goplani. SPDX-License-Identifier: MIT. See third_party/agent-deck/LICENSE.
package config

import "strings"

func NormalizeFooter(value string) string {
	switch value = strings.ToLower(strings.TrimSpace(value)); value {
	case "curated", "compact", "minimal":
		return value
	default:
		return "full"
	}
}
