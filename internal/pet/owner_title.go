package pet

import (
	"fmt"
	"strings"
)

const WorkbenchCaption = "Agent Session Desk"

// WindowMarker is transient launch identity, never a Session or Run target.
func WindowMarker(token string) (string, error) {
	if len(token) != 32 || strings.IndexFunc(token, func(r rune) bool { return !(r >= '0' && r <= '9' || r >= 'a' && r <= 'f') }) >= 0 {
		return "", fmt.Errorf("invalid workbench window token")
	}
	return "ASD Workbench " + token, nil
}
