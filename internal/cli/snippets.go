package cli

import (
	"context"
	"path/filepath"

	"github.com/liuhui-code/agent-session-desk/internal/session"
	"github.com/liuhui-code/agent-session-desk/internal/snippets"
	"github.com/liuhui-code/agent-session-desk/internal/ui"
)

type snippetDirectory interface {
	CurrentDirectory(context.Context, session.Session) (string, error)
}

func withSnippets(options ui.Options, source snippetDirectory, metadata string) ui.Options {
	global := filepath.Join(filepath.Dir(metadata), "snippets.toml")
	options.LoadSnippets = func(ctx context.Context, target session.Session) (snippets.Library, error) {
		dir, err := source.CurrentDirectory(ctx, target)
		if err != nil {
			dir = ""
		} // Explicit Global-only UI; never guess another directory.
		if err = ctx.Err(); err != nil {
			return snippets.Library{}, err
		}
		return snippets.Load(ctx, global, dir)
	}
	return options
}
