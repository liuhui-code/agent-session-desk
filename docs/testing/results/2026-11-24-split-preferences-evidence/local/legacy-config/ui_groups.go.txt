package config

import (
	"encoding/hex"
	"fmt"
	"sort"
	"strings"
)

// A sorted scalar keeps UISettings a comparable whole-document save snapshot.
// Only stable custom identities are stored; default has its existing bool.
func normalizeGroupFolds(value string) (string, error) {
	if value == "" {
		return "", nil
	}
	ids := strings.Split(value, ",")
	if len(ids) > 255 {
		return "", fmt.Errorf("折叠组数量超过上限")
	}
	seen := map[string]bool{}
	result := make([]string, 0, len(ids))
	for _, id := range ids {
		if len(id) != 32 || strings.ToLower(id) != id {
			return "", fmt.Errorf("折叠组身份无效")
		}
		if _, err := hex.DecodeString(id); err != nil {
			return "", fmt.Errorf("折叠组身份无效")
		}
		if !seen[id] {
			result = append(result, id)
			seen[id] = true
		}
	}
	sort.Strings(result)
	return strings.Join(result, ","), nil
}
