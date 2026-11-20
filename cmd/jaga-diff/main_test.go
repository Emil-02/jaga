package main

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"

	"github.com/Emil-02/jaga/internal/check"
	"github.com/Emil-02/jaga/internal/config"
	"github.com/Emil-02/jaga/internal/report"
)

func payload(score int, name, file string, line int, message string, sev config.Severity) report.Payload {
	results := []check.Result{}
	if name != "" {
		results = append(results, check.Result{
			Name:     name,
			Status:   "warn",
			Severity: sev,
			Findings: []check.Finding{{File: file, Line: line, Message: message, Severity: sev}},
		})
	}
	return report.Payload{Score: score, Results: results}
}

func TestCompare(t *testing.T) {
	before := payload(88, "secret", "a.js", 1, "lama", config.SeverityCritical)
	after := payload(76, "pii", "b.txt", 2, "baru", config.SeverityHigh)
	d := compare(before, after)
	if d.Before != 88 || d.After != 76 {
		t.Fatalf("skor = %d → %d", d.Before, d.After)
	}
	if len(d.New) != 1 || d.New[0].Check != "pii" {
		t.Fatalf("new = %+v", d.New)
	}
	if len(d.Resolved) != 1 || d.Resolved[0].Check != "secret" {
		t.Fatalf("resolved = %+v", d.Resolved)
	}
}

func TestCompareNoChanges(t *testing.T) {
	p := payload(90, "secret", "a.js", 1, "tetap", config.SeverityCritical)
	d := compare(p, p)
	if len(d.New) != 0 || len(d.Resolved) != 0 {
		t.Fatalf("diff = %+v", d)
	}
	if regressed(d, config.SeverityHigh) {
		t.Fatal("tanpa perubahan tidak boleh regresi")
	}
}

func TestRegressed(t *testing.T) {
	before := payload(90, "", "", 0, "", config.SeverityLow)
	cases := []struct {
		name  string
		after report.Payload
		want  bool
	}{
		{"skor turun", payload(80, "", "", 0, "", config.SeverityLow), true},
		{"temuan high baru", payload(90, "pii", "x", 1, "NIK", config.SeverityHigh), true},
		{"temuan low baru", payload(90, "hygiene", "x", 1, "LICENSE", config.SeverityLow), false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := regressed(compare(before, tc.after), config.SeverityHigh); got != tc.want {
				t.Fatalf("regressed = %v, mau %v", got, tc.want)
			}
		})
	}
}

func TestRenderMarkdown(t *testing.T) {
	before := payload(88, "secret", "a.js", 1, "lama", config.SeverityCritical)
	after := payload(76, "pii", "b.txt", 2, "baru", config.SeverityHigh)
	var buf bytes.Buffer
	if err := render(&buf, compare(before, after), "markdown"); err != nil {
		t.Fatal(err)
	}
	out := buf.String()
	for _, want := range []string{"**Skor: 88 → 76**", "turun 12", "Temuan baru", "`b.txt:2`", "Temuan selesai"} {
		if !strings.Contains(out, want) {
			t.Fatalf("markdown tidak mengandung %q:\n%s", want, out)
		}
	}
}

func TestRenderJSON(t *testing.T) {
	var buf bytes.Buffer
	if err := render(&buf, compare(payload(90, "", "", 0, "", config.SeverityLow), payload(90, "", "", 0, "", config.SeverityLow)), "json"); err != nil {
		t.Fatal(err)
	}
	var decoded map[string]any
	if err := json.Unmarshal(buf.Bytes(), &decoded); err != nil {
		t.Fatal(err)
	}
	if decoded["before"] != float64(90) || decoded["after"] != float64(90) {
		t.Fatalf("json = %v", decoded)
	}
}

func TestParseArgs(t *testing.T) {
	opts, err := parseArgs([]string{"before.json", "after.json", "--format", "markdown", "--fail-on", "critical", "--strict"})
	if err != nil {
		t.Fatal(err)
	}
	if opts.before != "before.json" || opts.after != "after.json" {
		t.Fatalf("paths = %q %q", opts.before, opts.after)
	}
	if opts.format != "markdown" || opts.failOn != "critical" || !opts.strict {
		t.Fatalf("opts = %+v", opts)
	}
	if _, err := parseArgs([]string{"before.json"}); err == nil {
		t.Fatal("satu file harusnya error")
	}
}

func TestLoad(t *testing.T) {
	p, err := load("../../jaga-report.json")
	if err != nil {
		t.Fatal(err)
	}
	if p.Score != 100 {
		t.Fatalf("skor = %d", p.Score)
	}
	if _, err := load("tidak-ada.json"); err == nil {
		t.Fatal("file tidak ada harus error")
	}
}
