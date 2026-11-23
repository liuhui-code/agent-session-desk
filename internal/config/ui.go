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
	Footer                string `toml:"footer"`
	Theme                 string `toml:"theme"`
	SidebarPercent        int    `toml:"sidebar_percent"`
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
	if (doc.Version < 1 || doc.Version > 7) || len(meta.Undecoded()) != 0 {
		return defaults, fmt.Errorf("界面设置版本或字段不受支持")
	}
	for _, key := range meta.Keys() {
		allowed := len(key) == 1 && (key[0] == "schema_version" || key[0] == "sidebar_density")
		if doc.Version >= 2 && len(key) == 1 && (key[0] == "sidebar_mode" || key[0] == "default_group_collapsed") {
			allowed = true
		}
		if doc.Version >= 3 && len(key) == 1 && key[0] == "collapsed_groups" {
			allowed = true
		}
		if doc.Version >= 4 && len(key) == 1 && key[0] == "sidebar_percent" {
			allowed = true
		}
		if doc.Version >= 5 && len(key) == 1 && key[0] == "theme" {
			allowed = true
		}
		if doc.Version >= 7 && len(key) == 1 && key[0] == "footer" {
			allowed = true
		}
		if !allowed {
			return defaults, fmt.Errorf("界面设置版本或字段不受支持")
		}
	}
	if !ValidSidebarPercent(doc.SidebarPercent) {
		return defaults, fmt.Errorf("侧栏比例必须为自动（0）或 10–90 的整数")
	}
	if !ValidUITheme(doc.Theme) || (meta.IsDefined("theme") && doc.Theme == "") || (doc.Version < 6 && doc.Theme == "system") {
		return defaults, fmt.Errorf("界面主题或对应版本不受支持")
	}
	doc.SidebarDensity = NormalizeSidebarDensity(doc.SidebarDensity)
	if doc.Footer != "" {
		doc.Footer = NormalizeFooter(doc.Footer)
	}
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
	if !ValidUITheme(settings.Theme) {
		return fmt.Errorf("界面主题必须为 dark、light 或 system")
	}
	if !ValidSidebarPercent(settings.SidebarPercent) {
		return fmt.Errorf("侧栏比例必须为自动（0）或 10–90 的整数")
	}
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
	if _, err = fmt.Fprintf(temp, "schema_version = 7\nsidebar_density = %q\nsidebar_mode = %q\ndefault_group_collapsed = %t\ncollapsed_groups = %q\nsidebar_percent = %d\n", normalized, mode, settings.DefaultGroupCollapsed, folds, settings.SidebarPercent); err != nil {
		return err
	}
	if settings.Theme != "" {
		if _, err = fmt.Fprintf(temp, "theme = %q\n", settings.Theme); err != nil {
			return err
		}
	}
	if settings.Footer != "" {
		if _, err = fmt.Fprintf(temp, "footer = %q\n", NormalizeFooter(settings.Footer)); err != nil {
			return err
		}
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

// ValidSidebarPercent is the contract introduced in v4, not a rendering size.
// Zero retains focus-dependent automatic geometry; manual values are bounded.
func ValidSidebarPercent(value int) bool {
	return value == 0 || (value >= 10 && value <= 90)
}

// Empty is the internal/default dark preference, serialized by omitting theme.
func ValidUITheme(value string) bool {
	return value == "" || value == "dark" || value == "light" || value == "system"
}
