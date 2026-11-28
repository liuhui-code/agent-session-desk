package updater

import (
	"github.com/liuhui-code/agent-session-desk/internal/backend"
	"runtime"
	"runtime/debug"
)

func CurrentCompatibility() Compatibility {
	return Compatibility{Metadata: 11, UI: 7, Launch: 1, Keybindings: 1, RecentDirs: 1, Recovery: 1, Runtime: backend.TerminalBundleFingerprints()}
}
func CurrentProbe() Probe {
	p := Probe{Protocol: 1, Platform: runtime.GOOS + "/" + runtime.GOARCH, Compatibility: CurrentCompatibility()}
	if b, ok := debug.ReadBuildInfo(); ok {
		for _, s := range b.Settings {
			switch s.Key {
			case "vcs.revision":
				p.Revision = s.Value
			case "vcs.modified":
				p.Dirty = s.Value == "true"
			}
		}
	}
	return p
}
