package check

import (
	"context"
	"regexp"
	"strings"

	"github.com/Emil-02/jaga/internal/config"
)

var (
	nikRE   = regexp.MustCompile(`\b[0-9]{16}\b`)
	npwpRE  = regexp.MustCompile(`\b[0-9]{2}\.[0-9]{3}\.[0-9]{3}\.[0-9]-[0-9]{3}\.[0-9]{3}\b`)
	phoneRE = regexp.MustCompile(`\b08[0-9]{8,11}\b`)
	bankRE  = regexp.MustCompile(`(?i)(rekening|no\.?\s*rek)\s*[:=]?\s*[0-9]{8,16}`)
)

type PII struct{}

func (PII) Name() string { return "pii" }

func (p PII) Run(ctx context.Context, root string, cfg config.Config) Result {
	cc := cfg.For(p.Name())
	res := NewResult(p.Name(), cc.Severity)
	exts := map[string]bool{"": true, ".go": true, ".js": true, ".ts": true, ".tsx": true, ".py": true, ".json": true, ".log": true, ".md": true, ".txt": true, ".csv": true, ".yml": true, ".yaml": true}
	err := ScanLines(ctx, root, cfg, exts, func(_ string, rel string, line int, text string) {
		for _, m := range nikRE.FindAllString(text, -1) {
			if validNIK(m) {
				res.Add(rel, line, "NIK/KK (16 digit): "+Mask(m))
			}
		}
		for _, m := range npwpRE.FindAllString(text, -1) {
			res.Add(rel, line, "NPWP: "+Mask(m))
		}
		for _, m := range phoneRE.FindAllString(text, -1) {
			res.Add(rel, line, "nomor HP: "+Mask(m))
		}
		for _, m := range bankRE.FindAllString(text, -1) {
			res.Add(rel, line, "nomor rekening: "+Mask(m))
		}
	})
	if err != nil {
		res.Error = err.Error()
	}
	return res
}

func validNIK(s string) bool {
	if len(s) != 16 || strings.Trim(s, "0") == "" {
		return false
	}
	first := 0
	for _, r := range s[:2] {
		first = first*10 + int(r-'0')
	}
	return first >= 1 && first <= 71
}
