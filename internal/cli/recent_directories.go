package cli

import (
	"context"
	"crypto/sha256"
	"fmt"
	"io"
	"path/filepath"

	"github.com/liuhui-code/agent-session-desk/internal/backend"
	"github.com/liuhui-code/agent-session-desk/internal/config"
)

func recentDirectoriesPath(opts backend.Options, dir string) (string, error) {
	metadata, scope, err := metadataPath(opts, dir)
	if err != nil {
		return "", err
	}
	return filepath.Join(filepath.Dir(metadata), fmt.Sprintf("recent-dirs-%x.toml", sha256.Sum256([]byte(scope)))), nil
}
func rememberLaunchDirectory(ctx context.Context, opts backend.Options, stateDir, dir string, warnings io.Writer) {
	if dir == "" {
		return
	}
	path, err := recentDirectoriesPath(opts, stateDir)
	if err == nil {
		err = config.RememberDirs(ctx, path, []string{dir})
	}
	if err != nil {
		fmt.Fprintln(warnings, "asd: 终端已启动；最近目录未保存：", safeError(err.Error()))
	}
}
