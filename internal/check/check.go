package check

import (
	"bufio"
	"context"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/Emil-02/jaga/internal/config"
)

type Finding struct {
	File     string          `json:"file"`
	Line     int             `json:"line"`
	Message  string          `json:"message"`
	Severity config.Severity `json:"severity"`
}

type Result struct {
	Name     string          `json:"name"`
	Status   string          `json:"status"`
	Severity config.Severity `json:"severity"`
	Findings []Finding       `json:"findings"`
	Error    string          `json:"error,omitempty"`
}

type Check interface {
	Name() string
	Run(ctx context.Context, root string, cfg config.Config) Result
}

func NewResult(name string, sev config.Severity) Result {
	return Result{Name: name, Status: "pass", Severity: sev, Findings: []Finding{}}
}

func (r *Result) Add(file string, line int, msg string) {
	r.Findings = append(r.Findings, Finding{File: file, Line: line, Message: msg, Severity: r.Severity})
	r.Status = "warn"
}

func Mask(s string) string {
	s = strings.TrimSpace(s)
	if len(s) <= 8 {
		return strings.Repeat("*", len(s))
	}
	return s[:4] + "…" + s[len(s)-4:]
}

type ScanFunc func(path string, rel string, line int, text string)

func ScanLines(ctx context.Context, root string, cfg config.Config, exts map[string]bool, fn ScanFunc) error {
	return filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if ctx.Err() != nil {
			return ctx.Err()
		}
		if d.IsDir() {
			if path != root && (cfg.IsExcluded(d.Name()) || strings.HasPrefix(d.Name(), ".") && d.Name() != ".") {
				return filepath.SkipDir
			}
			return nil
		}
		if strings.HasSuffix(d.Name(), "_test.go") {
			return nil
		}
		if len(exts) > 0 && !exts[strings.ToLower(filepath.Ext(path))] {
			return nil
		}
		info, err := d.Info()
		if err != nil || info.Size() > 2<<20 {
			return nil
		}
		f, err := os.Open(path)
		if err != nil {
			return nil
		}

		rel, err := filepath.Rel(root, path)
		if err != nil {
			rel = path
		}
		scanner := bufio.NewScanner(f)
		line := 0
		for scanner.Scan() {
			line++
			fn(path, rel, line, scanner.Text())
		}
		f.Close()
		return nil
	})
}
