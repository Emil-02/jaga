package check

import (
	"context"
	"regexp"

	"github.com/Emil-02/jaga/internal/config"
)

type secretRule struct {
	name string
	re   *regexp.Regexp
}

var secretRules = []secretRule{
	{"AWS access key", regexp.MustCompile(`AKIA[0-9A-Z]{16}`)},
	{"GitHub token", regexp.MustCompile(`gh[pousr]_[A-Za-z0-9_]{36,255}`)},
	{"OpenAI key", regexp.MustCompile(`sk-(proj-)?[A-Za-z0-9_-]{20,}`)},
	{"Slack token", regexp.MustCompile(`xox[baprs]-[A-Za-z0-9-]{10,}`)},
	{"Stripe live key", regexp.MustCompile(`sk_live_[A-Za-z0-9]{16,}`)},
	{"private key", regexp.MustCompile(`-----BEGIN (RSA |EC |OPENSSH |DSA )?PRIVATE KEY-----`)},
	{"generic credential", regexp.MustCompile(`(?i)(api[_-]?key|secret|token|password)\s*[:=]\s*["']?[A-Za-z0-9_\-]{16,}`)},
}

type Secret struct{}

func (Secret) Name() string { return "secret" }

func (s Secret) Run(ctx context.Context, root string, cfg config.Config) Result {
	cc := cfg.For(s.Name())
	res := NewResult(s.Name(), cc.Severity)
	exts := map[string]bool{
		".go": true, ".js": true, ".jsx": true, ".ts": true, ".tsx": true,
		".py": true, ".rb": true, ".php": true, ".java": true, ".env": true,
		".yml": true, ".yaml": true, ".json": true, ".toml": true, ".sh": true,
		".md": true, ".txt": true, ".example": true,
	}
	err := ScanLines(ctx, root, cfg, exts, func(_ string, rel string, line int, text string) {
		for _, rule := range secretRules {
			if m := rule.re.FindString(text); m != "" {
				res.Add(rel, line, rule.name+": "+Mask(m))
			}
		}
	})
	if err != nil {
		res.Error = err.Error()
	}
	return res
}
