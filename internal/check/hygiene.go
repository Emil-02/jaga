package check

import (
	"context"
	"os"
	"path/filepath"

	"github.com/Emil-02/jaga/internal/config"
)

type Hygiene struct{}

func (Hygiene) Name() string { return "hygiene" }

func (h Hygiene) Run(ctx context.Context, root string, cfg config.Config) Result {
	cc := cfg.For(h.Name())
	res := NewResult(h.Name(), cc.Severity)

	requirements := []struct {
		label string
		find  func() bool
	}{
		{"LICENSE", func() bool { return existsAny(root, "LICENSE", "LICENSE.md", "LICENSE.txt", "COPYING") }},
		{"README", func() bool { return existsAny(root, "README.md", "README", "README.rst", "README.txt") }},
		{"CI workflow", func() bool {
			matches, _ := filepath.Glob(filepath.Join(root, ".github", "workflows", "*.yml"))
			if len(matches) == 0 {
				matches, _ = filepath.Glob(filepath.Join(root, ".github", "workflows", "*.yaml"))
			}
			return len(matches) > 0
		}},
		{".gitignore", func() bool { return existsAny(root, ".gitignore") }},
	}

	for _, req := range requirements {
		if ctx.Err() != nil {
			break
		}
		if !req.find() {
			res.Add(".", 0, req.label+" tidak ditemukan")
		}
	}
	return res
}

func existsAny(root string, names ...string) bool {
	for _, name := range names {
		if _, err := os.Stat(filepath.Join(root, name)); err == nil {
			return true
		}
	}
	return false
}
