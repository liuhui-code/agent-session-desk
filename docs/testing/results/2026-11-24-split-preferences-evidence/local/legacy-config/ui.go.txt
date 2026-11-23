package config

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/BurntSushi/toml"
)

const maxUIPreferenceBytes = 16 * 1024 // Defensive bound for UI preferences, not a resource budget.

type UISettings struct {
	SidebarDensity        string `toml:"sidebar_density"`
	SidebarMode           string `toml:"sidebar_mode"`
	DefaultGroupCollapsed bool   `toml:"default_group_collapsed"`
	CollapsedGroups       string `toml:"collapsed_groups"`
}

func NormalizeSidebarMode(value string) string {
	if value == "flat" {
		return "flat"
	}
	return "grouped"
}

type uiDocument struct {
	Version int `toml:"schema_version"`
	UISettings
}

func NormalizeSidebarDensity(value string) string {
	switch value = strings.ToLower(strings.TrimSpace(value)); value {
	case "full", "compact", "minimal", "auto":
		return value
	default:
		return "compact"
	}
}

// LoadUI reads only the dedicated preference document. Launch configuration is
// never edited. Missing files use upstream's compact default.
func LoadUI(path string) (UISettings, error) {
	defaults := UISettings{SidebarDensity: "compact", SidebarMode: "grouped"}
	info, err := os.Lstat(path)
	if errors.Is(err, os.ErrNotExist) {
		return defaults, nil
	}
	if err != nil {
		return defaults, err
	}
	if !info.Mode().IsRegular() {
		return defaults, fmt.Errorf("界面设置必须是普通文件")
	}
	f, err := os.Open(path)
	if err != nil {
		return defaults, err
	}
	defer f.Close()
	data, err := io.ReadAll(io.LimitReader(f, maxUIPreferenceBytes+1))
	if err != nil {
		return defaults, err
	}
	if len(data) > maxUIPreferenceBytes {
		return defaults, fmt.Errorf("界面设置超过读取上限")
	}
	var doc uiDocument
	meta, err := toml.Decode(string(data), &doc)
	if err != nil {
		return defaults, fmt.Errorf("界面设置 TOML 或字段类型无效")
	}
	if (doc.Version != 1 && doc.Version != 2 && doc.Version != 3) || len(meta.Undecoded()) != 0 {
		return defaults, fmt.Errorf("界面设置版本或字段不受支持")
	}
	for _, key := range meta.Keys() {
		allowed := len(key) == 1 && (key[0] == "schema_version" || key[0] == "sidebar_density")
		if doc.Version >= 2 && len(key) == 1 && (key[0] == "sidebar_mode" || key[0] == "default_group_collapsed") {
			allowed = true
		}
		if doc.Version == 3 && len(key) == 1 && key[0] == "collapsed_groups" {
			allowed = true
		}
		if !allowed {
			return defaults, fmt.Errorf("界面设置版本或字段不受支持")
		}
	}
	doc.SidebarDensity = NormalizeSidebarDensity(doc.SidebarDensity)
	doc.SidebarMode = NormalizeSidebarMode(doc.SidebarMode)
	doc.CollapsedGroups, err = normalizeGroupFolds(doc.CollapsedGroups)
	if err != nil {
		return defaults, err
	}
	return doc.UISettings, nil
}

// SaveUI publishes a synced, validated same-directory temporary document by
// rename. Concurrent valid preference writers use last successful writer wins;
// this is not a transaction with metadata or a power-loss durability claim.
func SaveUI(path string, settings UISettings) error {
	folds, err := normalizeGroupFolds(settings.CollapsedGroups)
	if err != nil {
		return err
	}
	normalized := NormalizeSidebarDensity(settings.SidebarDensity)
	if normalized != settings.SidebarDensity {
		return fmt.Errorf("界面密度无效")
	}
	mode := NormalizeSidebarMode(settings.SidebarMode)
	if settings.SidebarMode != "" && mode != settings.SidebarMode {
		return fmt.Errorf("侧栏模式无效")
	}
	if _, err := LoadUI(path); err != nil {
		return err
	}
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0700); err != nil {
		return err
	}
	temp, err := os.CreateTemp(dir, ".ui-*.tmp")
	if err != nil {
		return err
	}
	name := temp.Name()
	defer os.Remove(name)
	closeTemp := true
	defer func() {
		if closeTemp {
			_ = temp.Close()
		}
	}()
	if err = temp.Chmod(0600); err != nil {
		return err
	}
	if _, err = fmt.Fprintf(temp, "schema_version = 3\nsidebar_density = %q\nsidebar_mode = %q\ndefault_group_collapsed = %t\ncollapsed_groups = %q\n", normalized, mode, settings.DefaultGroupCollapsed, folds); err != nil {
		return err
	}
	if err = temp.Sync(); err != nil {
		return err
	}
	if err = temp.Close(); err != nil {
		return err
	}
	closeTemp = false
	if _, err = LoadUI(name); err != nil {
		return err
	}
	if _, err = LoadUI(path); err != nil {
		return err
	}
	return os.Rename(name, path)
}
