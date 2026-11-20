package main

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/Emil-02/jaga/internal/check"
	"github.com/Emil-02/jaga/internal/config"
	"github.com/Emil-02/jaga/internal/report"
)

type finding struct {
	Check    string          `json:"check"`
	File     string          `json:"file"`
	Line     int             `json:"line"`
	Message  string          `json:"message"`
	Severity config.Severity `json:"severity"`
}

type diff struct {
	Before   int       `json:"before"`
	After    int       `json:"after"`
	New      []finding `json:"new"`
	Resolved []finding `json:"resolved"`
}

func key(f finding) string {
	return fmt.Sprintf("%s|%s|%d|%s", f.Check, f.File, f.Line, f.Message)
}

func flatten(p report.Payload) []finding {
	out := []finding{}
	for _, r := range p.Results {
		for _, f := range r.Findings {
			out = append(out, finding{
				Check:    r.Name,
				File:     f.File,
				Line:     f.Line,
				Message:  f.Message,
				Severity: f.Severity,
			})
		}
	}
	return out
}

func compare(before, after report.Payload) diff {
	old := map[string]finding{}
	for _, f := range flatten(before) {
		old[key(f)] = f
	}
	next := map[string]finding{}
	for _, f := range flatten(after) {
		next[key(f)] = f
	}
	d := diff{Before: before.Score, After: after.Score, New: []finding{}, Resolved: []finding{}}
	for k, f := range next {
		if _, ok := old[k]; !ok {
			d.New = append(d.New, f)
		}
	}
	for k, f := range old {
		if _, ok := next[k]; !ok {
			d.Resolved = append(d.Resolved, f)
		}
	}
	sortFindings(d.New)
	sortFindings(d.Resolved)
	return d
}

func sortFindings(items []finding) {
	for i := 1; i < len(items); i++ {
		for j := i; j > 0 && key(items[j]) < key(items[j-1]); j-- {
			items[j], items[j-1] = items[j-1], items[j]
		}
	}
}

func regressed(d diff, failOn config.Severity) bool {
	if d.After < d.Before {
		return true
	}
	for _, f := range d.New {
		if f.Severity.Rank() >= failOn.Rank() {
			return true
		}
	}
	return false
}

func load(path string) (report.Payload, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return report.Payload{}, err
	}
	var p report.Payload
	if err := json.Unmarshal(data, &p); err != nil {
		return report.Payload{}, fmt.Errorf("%s: JSON tidak valid: %w", path, err)
	}
	if p.Results == nil {
		p.Results = []check.Result{}
	}
	return p, nil
}

func render(w io.Writer, d diff, format string) error {
	switch format {
	case "text":
		return renderText(w, d)
	case "json":
		return json.NewEncoder(w).Encode(d)
	case "markdown":
		return renderMarkdown(w, d)
	default:
		return fmt.Errorf("format tidak dikenal: %s", format)
	}
}

func renderMarkdown(w io.Writer, d diff) error {
	var b strings.Builder
	delta := d.After - d.Before
	b.WriteString("### jaga — perubahan skor\n\n")
	if delta == 0 && len(d.New) == 0 && len(d.Resolved) == 0 {
		b.WriteString("Tidak ada perubahan skor atau temuan.\n")
	} else {
		trend := "tetap"
		if delta > 0 {
			trend = fmt.Sprintf("naik %d", delta)
		} else if delta < 0 {
			trend = fmt.Sprintf("turun %d", -delta)
		}
		fmt.Fprintf(&b, "**Skor: %d → %d** (%s)\n\n", d.Before, d.After, trend)
		fmt.Fprintf(&b, "| Perubahan | Jumlah |\n|---|---:|\n| Temuan baru | %d |\n| Temuan selesai | %d |\n", len(d.New), len(d.Resolved))
		if len(d.New) > 0 {
			b.WriteString("\n#### Temuan baru\n\n")
			for _, f := range d.New {
				fmt.Fprintf(&b, "- **%s** — `%s` — %s\n", f.Check, location(f), f.Message)
			}
		}
		if len(d.Resolved) > 0 {
			b.WriteString("\n#### Temuan selesai\n\n")
			for _, f := range d.Resolved {
				fmt.Fprintf(&b, "- **%s** — `%s` — %s\n", f.Check, location(f), f.Message)
			}
		}
	}
	_, err := io.WriteString(w, b.String())
	return err
}

func renderText(w io.Writer, d diff) error {
	var b strings.Builder
	fmt.Fprintf(&b, "jaga diff — skor %d → %d\n", d.Before, d.After)
	fmt.Fprintf(&b, "temuan baru: %d | temuan selesai: %d\n", len(d.New), len(d.Resolved))
	for _, f := range d.New {
		fmt.Fprintf(&b, "+ %s: %s — %s\n", f.Check, location(f), f.Message)
	}
	for _, f := range d.Resolved {
		fmt.Fprintf(&b, "- %s: %s — %s\n", f.Check, location(f), f.Message)
	}
	_, err := io.WriteString(w, b.String())
	return err
}

func location(f finding) string {
	if f.Line > 0 {
		return fmt.Sprintf("%s:%d", f.File, f.Line)
	}
	return f.File
}

type args struct {
	before string
	after  string
	format string
	failOn string
	strict bool
}

func parseArgs(argv []string) (args, error) {
	opts := args{format: "markdown", failOn: "high"}
	paths := []string{}
	for i := 0; i < len(argv); i++ {
		a := argv[i]
		next := func() (string, error) {
			if i+1 >= len(argv) {
				return "", fmt.Errorf("nilai untuk %s tidak ada", a)
			}
			i++
			return argv[i], nil
		}
		switch a {
		case "--format", "-format":
			v, err := next()
			if err != nil {
				return opts, err
			}
			opts.format = v
		case "--fail-on":
			v, err := next()
			if err != nil {
				return opts, err
			}
			opts.failOn = v
		case "--strict":
			opts.strict = true
		default:
			if strings.HasPrefix(a, "-") {
				return opts, fmt.Errorf("argumen tidak dikenal: %s", a)
			}
			paths = append(paths, a)
		}
	}
	if len(paths) != 2 {
		return opts, fmt.Errorf("butuh dua file laporan: before.json after.json")
	}
	opts.before, opts.after = paths[0], paths[1]
	return opts, nil
}

func main() {
	opts, err := parseArgs(os.Args[1:])
	if err != nil {
		fmt.Fprintln(os.Stderr, "jaga-diff:", err)
		fmt.Fprintln(os.Stderr, "pemakaian: jaga-diff before.json after.json [--format markdown|text|json] [--strict]")
		os.Exit(2)
	}
	sev, ok := config.ParseSeverity(opts.failOn)
	if !ok {
		fmt.Fprintln(os.Stderr, "jaga-diff: fail_on tidak valid:", opts.failOn)
		os.Exit(2)
	}

	before, err := load(opts.before)
	if err != nil {
		fmt.Fprintln(os.Stderr, "jaga-diff:", err)
		os.Exit(2)
	}
	after, err := load(opts.after)
	if err != nil {
		fmt.Fprintln(os.Stderr, "jaga-diff:", err)
		os.Exit(2)
	}
	d := compare(before, after)
	if err := render(os.Stdout, d, opts.format); err != nil {
		fmt.Fprintln(os.Stderr, "jaga-diff:", err)
		os.Exit(2)
	}
	if opts.strict && regressed(d, sev) {
		os.Exit(1)
	}
}
