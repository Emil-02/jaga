package runner

import (
	"context"
	"sort"
	"sync"

	"github.com/Emil-02/jaga/internal/check"
	"github.com/Emil-02/jaga/internal/config"
)

type Runner struct {
	checks []check.Check
}

func NewDefault() *Runner {
	return &Runner{checks: []check.Check{
		check.Secret{},
		check.PII{},
		check.Hygiene{},
	}}
}

func (r *Runner) Add(c check.Check) {
	r.checks = append(r.checks, c)
}

func (r *Runner) RunAll(ctx context.Context, root string, cfg config.Config) []check.Result {
	var wg sync.WaitGroup
	results := make(chan check.Result, len(r.checks))

	for _, c := range r.checks {
		if !cfg.For(c.Name()).Enabled {
			continue
		}
		wg.Add(1)
		go func(c check.Check) {
			defer wg.Done()
			results <- c.Run(ctx, root, cfg)
		}(c)
	}
	wg.Wait()
	close(results)

	out := make([]check.Result, 0, len(r.checks))
	for res := range results {
		out = append(out, res)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Name < out[j].Name })
	return out
}
