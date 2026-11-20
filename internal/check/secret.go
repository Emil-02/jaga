package check

import (
	"context"
	"encoding/base64"
	"regexp"
	"strings"

	"github.com/Emil-02/jaga/internal/config"
)

type secretRule struct {
	name     string
	re       *regexp.Regexp
	validate func(match string, lineText string) bool
}

var placeholderRE = regexp.MustCompile(`(?i)(x{6,}|\*{4,}|your[_-]?key|your[_-]?secret|your[_-]?token|your[_-]?api|example|placeholder|dummy|changeme|redacted|<[^>]+>|\.\.\.)`)

func isPlaceholder(s string) bool {
	return placeholderRE.MatchString(s)
}

func isSupabaseServiceRole(jwtStr string) bool {
	parts := strings.Split(jwtStr, ".")
	if len(parts) < 2 {
		return false
	}
	payload, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		payload, err = base64.URLEncoding.DecodeString(parts[1])
		if err != nil {
			return false
		}
	}
	pStr := string(payload)
	return strings.Contains(pStr, `"role":"service_role"`) ||
		strings.Contains(pStr, `"role": "service_role"`) ||
		strings.Contains(pStr, `"role":"supabase_admin"`) ||
		strings.Contains(pStr, `"role": "supabase_admin"`)
}

var secretRules = []secretRule{
	// Cloud & Version Control
	{name: "AWS access key", re: regexp.MustCompile(`\bAKIA[0-9A-Z]{16}\b`)},
	{name: "AWS secret access key", re: regexp.MustCompile(`(?i)\baws[_-]?secret[_-]?access[_-]?key\s*[:=]\s*["']?[A-Za-z0-9\/+=]{40}`)},
	{name: "GitHub token", re: regexp.MustCompile(`\bgh[pousr]_[A-Za-z0-9_]{36,255}\b`)},
	{name: "GitLab token", re: regexp.MustCompile(`\bglpat-[A-Za-z0-9\-]{20,}\b`)},
	{name: "Slack token", re: regexp.MustCompile(`\bxox[baprs]-[A-Za-z0-9-]{10,}\b`)},

	// AI & LLM Providers (Modern Vibe-Coding)
	{
		name: "OpenAI API key",
		re:   regexp.MustCompile(`\bsk-(?:proj-|svcacct-|admin-)?[A-Za-z0-9_-]{20,}\b`),
		validate: func(m string, _ string) bool {
			return !strings.HasPrefix(m, "sk-ant-") && !isPlaceholder(m)
		},
	},
	{name: "Anthropic API key", re: regexp.MustCompile(`\bsk-ant-(?:api[0-9]{2}-)?[A-Za-z0-9_-]{30,}\b`)},
	{name: "Google Gemini/Cloud API key", re: regexp.MustCompile(`\bAIzaSy[A-Za-z0-9_-]{33}\b`)},
	{name: "Groq API key", re: regexp.MustCompile(`\bgsk_[A-Za-z0-9]{48}\b`)},
	{name: "HuggingFace token", re: regexp.MustCompile(`\bhf_[A-Za-z0-9]{34}\b`)},
	{name: "Replicate API token", re: regexp.MustCompile(`\br8_[A-Za-z0-9]{37}\b`)},

	// Backend, Database & BaaS
	{
		name: "Supabase service_role key",
		re:   regexp.MustCompile(`\b(sb_secret_[A-Za-z0-9_-]{20,}|eyJ[A-Za-z0-9_-]{10,}\.eyJ[A-Za-z0-9_-]{20,}\.[A-Za-z0-9_-]{20,})\b`),
		validate: func(m string, _ string) bool {
			if strings.HasPrefix(m, "sb_secret_") {
				return !isPlaceholder(m)
			}
			return isSupabaseServiceRole(m)
		},
	},
	{name: "Resend API key", re: regexp.MustCompile(`\bre_[A-Za-z0-9]{24,}\b`)},
	{name: "Database URI with password", re: regexp.MustCompile(`(?i)\b(postgres|postgresql|mysql|mongodb|mongodb\+srv|redis|amqp|mssql):\/\/[A-Za-z0-9_\-\.\%]+:[^@\s"']+@[A-Za-z0-9_\-\.]+`)},

	// Indonesian Payment Gateways & Fintech
	{name: "Midtrans Server key", re: regexp.MustCompile(`\b(?:SB-)?Mid-server-[A-Za-z0-9_\-]{20,}\b`)},
	{name: "Xendit Secret key", re: regexp.MustCompile(`\bxnd_(?:development|production)_[A-Za-z0-9_\-]{20,}\b`)},

	// Payment & Security
	{name: "Stripe live key", re: regexp.MustCompile(`\b(?:sk|rk)_live_[A-Za-z0-9]{16,}\b`)},
	{name: "Stripe webhook secret", re: regexp.MustCompile(`\bwhsec_[A-Za-z0-9]{24,}\b`)},
	{name: "private key", re: regexp.MustCompile(`-----BEGIN (RSA |EC |OPENSSH |DSA )?PRIVATE KEY-----`)},

	// Vibe-Coding Frontend Leak Smell (Exposing private/secret variables in client env)
	{name: "Vibe-code frontend secret leak", re: regexp.MustCompile(`(?i)\b(NEXT_PUBLIC_|VITE_|REACT_APP_|PUBLIC_)[A-Za-z0-9_]*(SECRET|PRIVATE|SERVICE_ROLE|ADMIN_KEY)[A-Za-z0-9_]*\s*[:=]\s*["']?[^"'\s\n]{8,}`)},

	// Generic Credential Fallback
	{name: "generic credential", re: regexp.MustCompile(`(?i)(api[_-]?key|secret[_-]?key|auth[_-]?token|app[_-]?password)\s*[:=]\s*["']?[A-Za-z0-9_\-]{16,}`)},
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
		".md": true, ".txt": true, ".example": true, ".sql": true,
	}
	err := ScanLines(ctx, root, cfg, exts, func(_ string, rel string, line int, text string) {
		for _, rule := range secretRules {
			if m := rule.re.FindString(text); m != "" {
				if rule.validate != nil {
					if !rule.validate(m, text) {
						continue
					}
				} else if isPlaceholder(m) {
					continue
				}
				res.Add(rel, line, rule.name+": "+Mask(m))
			}
		}
	})
	if err != nil {
		res.Error = err.Error()
	}
	return res
}
