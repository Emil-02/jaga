package setup

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/Emil-02/jaga/internal/config"
	"github.com/Emil-02/jaga/internal/report"
	"github.com/Emil-02/jaga/internal/runner"
	"github.com/Emil-02/jaga/internal/score"
)

const configTemplate = `version: 1
exclude:
  - node_modules
  - dist
  - vendor
  - .git
checks:
  secret:
    enabled: true
    severity: critical
  pii:
    enabled: true
    severity: high
  hygiene:
    enabled: true
    severity: low
fail_on: high
`

const workflowTemplate = `name: jaga

on:
  push:
    branches: [main]
  pull_request:

permissions:
  contents: write
  pull-requests: write

jobs:
  pr-gate:
    if: github.event_name == 'pull_request'
    uses: Emil-02/jaga/.github/workflows/jaga-pr.yml@main
    with:
      strict: true
      fail-on: high

  report:
    if: github.event_name == 'push'
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
        with:
          token: ${{ secrets.JAGA_PAT }}
      - uses: Emil-02/jaga@main
        with:
          report: "true"
          badge: "true"
      - name: Commit laporan
        run: |
          git config user.name "%s"
          git config user.email "%s"
          git add jaga-report.json badge.svg
          if ! git diff --cached --quiet; then
            git commit -m "chore(jaga): update report"
            git push
          fi
`

type Options struct {
	Root        string
	Force       bool
	SkipScan    bool
	AuthorName  string
	AuthorEmail string
}

type Result struct {
	Created []string
	Skipped []string
}

func Run(opts Options) (Result, error) {
	result := Result{Created: []string{}, Skipped: []string{}}
	root, err := filepath.Abs(opts.Root)
	if err != nil {
		return result, err
	}
	info, err := os.Stat(root)
	if err != nil || !info.IsDir() {
		return result, fmt.Errorf("bukan direktori: %s", opts.Root)
	}
	if err := validateAuthor(opts.AuthorName, opts.AuthorEmail); err != nil {
		return result, err
	}

	if err := writeIfMissing(filepath.Join(root, ".jaga.yml"), configTemplate, ".jaga.yml", opts.Force, &result); err != nil {
		return result, err
	}

	workflow := fmt.Sprintf(workflowTemplate, opts.AuthorName, opts.AuthorEmail)
	workflowPath := filepath.Join(root, ".github", "workflows", "jaga.yml")
	if err := os.MkdirAll(filepath.Dir(workflowPath), 0o755); err != nil {
		return result, err
	}
	if err := writeIfMissing(workflowPath, workflow, ".github/workflows/jaga.yml", opts.Force, &result); err != nil {
		return result, err
	}

	if err := ensureBadge(root, &result); err != nil {
		return result, err
	}

	if !opts.SkipScan {
		if err := writeReport(root, opts.Force, &result); err != nil {
			return result, err
		}
	}
	return result, nil
}

func validateAuthor(name, email string) error {
	for _, value := range []struct {
		label string
		text  string
	}{{"author-name", name}, {"author-email", email}} {
		if strings.TrimSpace(value.text) == "" {
			return fmt.Errorf("%s wajib diisi", value.label)
		}
		if len(value.text) > 120 || strings.ContainsAny(value.text, "\r\n\"") || strings.Contains(value.text, "${{") {
			return fmt.Errorf("%s mengandung karakter yang tidak diizinkan", value.label)
		}
	}
	return nil
}

func writeIfMissing(path, content, label string, force bool, result *Result) error {
	if _, err := os.Stat(path); err == nil && !force {
		result.Skipped = append(result.Skipped, label+" (sudah ada)")
		return nil
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		return err
	}
	result.Created = append(result.Created, label)
	return nil
}

func ensureBadge(root string, result *Result) error {
	owner, repo, err := repoSlug(root)
	if err != nil {
		result.Skipped = append(result.Skipped, "badge README (remote GitHub tidak ditemukan)")
		return nil
	}

	path := filepath.Join(root, "README.md")
	data, err := os.ReadFile(path)
	badge := fmt.Sprintf(
		"[![jaga](https://jaga-card.vercel.app/api/badge?owner=%s&repo=%s)](https://jaga-card.vercel.app/r/%s/%s)",
		owner, repo, owner, repo,
	)
	if os.IsNotExist(err) {
		content := "# " + repo + "\n\n" + badge + "\n"
		if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
			return err
		}
		result.Created = append(result.Created, "README.md (badge)")
		return nil
	}
	if err != nil {
		return err
	}
	text := string(data)
	if strings.Contains(text, "jaga-card.vercel.app/api/badge") {
		result.Skipped = append(result.Skipped, "badge README (sudah ada)")
		return nil
	}
	lines := strings.SplitN(text, "\n", 2)
	updated := lines[0] + "\n\n" + badge + "\n"
	if len(lines) > 1 {
		updated += lines[1]
	}
	if err := os.WriteFile(path, []byte(updated), 0o644); err != nil {
		return err
	}
	result.Created = append(result.Created, "README.md (badge)")
	return nil
}

func repoSlug(root string) (string, string, error) {
	out, err := exec.Command("git", "-C", root, "remote", "get-url", "origin").Output()
	if err != nil {
		return "", "", err
	}
	s := strings.TrimSpace(string(out))
	if strings.HasPrefix(s, "git@github.com:") {
		s = strings.TrimPrefix(s, "git@github.com:")
	} else if i := strings.Index(s, "github.com/"); i >= 0 {
		s = s[i+len("github.com/"):]
	}
	s = strings.TrimSuffix(s, ".git")
	parts := strings.Split(s, "/")
	if len(parts) < 2 || parts[0] == "" || parts[1] == "" {
		return "", "", fmt.Errorf("remote origin bukan GitHub: %s", s)
	}
	return parts[0], parts[1], nil
}

func writeReport(root string, force bool, result *Result) error {
	path := filepath.Join(root, "jaga-report.json")
	if _, err := os.Stat(path); err == nil && !force {
		result.Skipped = append(result.Skipped, "jaga-report.json (sudah ada)")
		return nil
	}
	cfg, err := config.Load(config.Find(root))
	if err != nil {
		return err
	}
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()
	results := runner.NewDefault().RunAll(ctx, root, cfg)
	payload := report.Payload{
		Root:        root,
		Score:       score.Calculate(results),
		GeneratedAt: time.Now().UTC().Format(time.RFC3339),
		Results:     results,
	}
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	if err := report.Render(f, payload, "json"); err != nil {
		return err
	}
	result.Created = append(result.Created, "jaga-report.json")
	return nil
}
