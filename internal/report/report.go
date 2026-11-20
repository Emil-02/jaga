package report

import (
	"encoding/json"
	"fmt"
	"html"
	"io"
	"strings"

	"github.com/Emil-02/jaga/internal/check"
)

type Payload struct {
	Root        string         `json:"root"`
	Score       int            `json:"score"`
	GeneratedAt string         `json:"generated_at,omitempty"`
	Results     []check.Result `json:"results"`
}

func Render(w io.Writer, p Payload, format string) error {
	switch format {
	case "text":
		return renderText(w, p)
	case "json":
		return json.NewEncoder(w).Encode(p)
	case "sarif":
		return renderSARIF(w, p)
	case "badge":
		return renderBadge(w, p)
	case "html":
		return renderHTML(w, p)
	case "markdown", "md":
		return renderMarkdown(w, p)
	default:
		return fmt.Errorf("format tidak dikenal: %s", format)
	}
}

func renderText(w io.Writer, p Payload) error {
	var b strings.Builder
	fmt.Fprintf(&b, "jaga — %s\n", p.Root)
	b.WriteString(strings.Repeat("─", 56) + "\n")
	for _, r := range p.Results {
		mark := "✓"
		if r.Status == "warn" {
			mark = "!"
		}
		if r.Error != "" {
			mark = "✗"
		}
		fmt.Fprintf(&b, "%s %-10s %s\n", mark, r.Name, strings.ToUpper(string(r.Severity)))
		for _, f := range r.Findings {
			if f.Line > 0 {
				fmt.Fprintf(&b, "    %s:%d — %s\n", f.File, f.Line, f.Message)
			} else {
				fmt.Fprintf(&b, "    %s — %s\n", f.File, f.Message)
			}
		}
		if r.Error != "" {
			fmt.Fprintf(&b, "    error: %s\n", r.Error)
		}
	}
	b.WriteString(strings.Repeat("─", 56) + "\n")
	fmt.Fprintf(&b, "Skor: %d/100\n", p.Score)
	_, err := io.WriteString(w, b.String())
	return err
}

type sarifTool struct {
	Driver struct {
		Name string `json:"name"`
	} `json:"driver"`
}

type sarifResult struct {
	RuleID  string `json:"ruleId"`
	Level   string `json:"level"`
	Message struct {
		Text string `json:"text"`
	} `json:"message"`
	Locations []struct {
		PhysicalLocation struct {
			ArtifactLocation struct {
				URI string `json:"uri"`
			} `json:"artifactLocation"`
			Region struct {
				StartLine int `json:"startLine"`
			} `json:"region"`
		} `json:"physicalLocation"`
	} `json:"locations"`
}

type sarifRun struct {
	Tool    sarifTool     `json:"tool"`
	Results []sarifResult `json:"results"`
}

type sarifReport struct {
	Version string     `json:"version"`
	Schema  string     `json:"$schema"`
	Runs    []sarifRun `json:"runs"`
}

func renderSARIF(w io.Writer, p Payload) error {
	run := sarifRun{Results: []sarifResult{}}
	run.Tool.Driver.Name = "jaga"
	for _, r := range p.Results {
		for _, f := range r.Findings {
			item := sarifResult{RuleID: r.Name, Level: level(string(r.Severity))}
			item.Message.Text = f.Message
			loc := struct {
				PhysicalLocation struct {
					ArtifactLocation struct {
						URI string `json:"uri"`
					} `json:"artifactLocation"`
					Region struct {
						StartLine int `json:"startLine"`
					} `json:"region"`
				} `json:"physicalLocation"`
			}{}
			loc.PhysicalLocation.ArtifactLocation.URI = f.File
			line := f.Line
			if line < 1 {
				line = 1
			}
			loc.PhysicalLocation.Region.StartLine = line
			item.Locations = append(item.Locations, loc)
			run.Results = append(run.Results, item)
		}
	}
	doc := sarifReport{
		Version: "2.1.0",
		Schema:  "https://json.schemastore.org/sarif-2.1.0.json",
		Runs:    []sarifRun{run},
	}
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(doc)
}

func level(sev string) string {
	switch sev {
	case "critical", "high":
		return "error"
	case "medium":
		return "warning"
	default:
		return "note"
	}
}

func renderBadge(w io.Writer, p Payload) error {
	color := "#4c1"
	switch {
	case p.Score < 50:
		color = "#e05d44"
	case p.Score < 80:
		color = "#dfb317"
	}
	svg := fmt.Sprintf(`<svg xmlns="http://www.w3.org/2000/svg" width="106" height="20" role="img" aria-label="jaga: %d/100">
  <linearGradient id="s" x2="0" y2="100%%"><stop offset="0" stop-color="#bbb" stop-opacity=".1"/><stop offset="1" stop-opacity=".1"/></linearGradient>
  <clipPath id="r"><rect width="106" height="20" rx="3" fill="#fff"/></clipPath>
  <g clip-path="url(#r)"><rect width="42" height="20" fill="#555"/><rect x="42" width="64" height="20" fill="%s"/><rect width="106" height="20" fill="url(#s)"/></g>
  <g fill="#fff" text-anchor="middle" font-family="Verdana,Geneva,DejaVu Sans,sans-serif" font-size="11">
    <text x="21" y="15">jaga</text><text x="74" y="15">%d/100</text>
  </g>
</svg>
`, p.Score, color, p.Score)
	_, err := io.WriteString(w, svg)
	return err
}

const htmlStyle = `
:root{color-scheme:dark;--bg:#0b1120;--card:#111c34;--line:#1e293b;--text:#e2e8f0;--muted:#94a3b8}
*{box-sizing:border-box}
body{margin:0;background:var(--bg);color:var(--text);font:15px/1.6 ui-sans-serif,system-ui,-apple-system,"Segoe UI",sans-serif}
main{max-width:860px;margin:0 auto;padding:40px 20px 64px}
header{display:flex;justify-content:space-between;align-items:baseline;gap:16px;border-bottom:1px solid var(--line);padding-bottom:16px;margin-bottom:32px}
.brand{font-size:22px;font-weight:700;letter-spacing:-.02em}
.brand span{font-size:18px}
.path{color:var(--muted);font:13px ui-monospace,SFMono-Regular,Menlo,monospace;word-break:break-all;text-align:right}
.hero{display:flex;gap:32px;align-items:center;margin-bottom:40px}
.ring{--score:0;--color:#4ade80;flex:none;width:132px;height:132px;border-radius:50%;background:conic-gradient(var(--color) calc(var(--score)*1%),var(--line) 0);display:grid;place-items:center}
.ring-inner{width:106px;height:106px;border-radius:50%;background:var(--bg);display:flex;flex-direction:column;align-items:center;justify-content:center}
.ring-inner strong{font-size:34px;line-height:1;color:var(--color)}
.ring-inner span{font-size:12px;color:var(--muted)}
h1{font-size:26px;margin:0 0 8px;letter-spacing:-.02em}
.hero-copy p{margin:0 0 20px;color:var(--muted)}
.stats{display:flex;gap:24px;flex-wrap:wrap}
.stat b{display:block;font-size:20px;line-height:1.2}
.stat span{font-size:12px;color:var(--muted);text-transform:uppercase;letter-spacing:.06em}
h2{font-size:14px;text-transform:uppercase;letter-spacing:.1em;color:var(--muted);margin:0 0 16px}
.card{background:var(--card);border:1px solid var(--line);border-radius:12px;padding:18px 20px;margin-bottom:14px}
.card-head{display:flex;align-items:center;gap:10px;flex-wrap:wrap}
.card-head h3{margin:0;font-size:16px;font-family:ui-monospace,SFMono-Regular,Menlo,monospace}
.pill{font-size:10px;font-weight:700;letter-spacing:.08em;text-transform:uppercase;padding:3px 8px;border-radius:999px;border:1px solid currentColor;opacity:.9}
.pill.low{color:#94a3b8}.pill.medium{color:#fbbf24}.pill.high{color:#fb923c}.pill.critical{color:#f87171}
.status{margin-left:auto;font-size:11px;color:var(--muted)}
.status.warn{color:#fbbf24}.status.error{color:#f87171}
.findings{list-style:none;margin:14px 0 0;padding:0}
.findings li{padding:8px 0;border-top:1px solid var(--line)}
.findings code{color:#7dd3fc;font:12px ui-monospace,SFMono-Regular,Menlo,monospace;margin-right:8px}
.findings li span{color:var(--text)}
.empty{margin:14px 0 0;color:var(--muted);font-size:13px}
footer{border-top:1px solid var(--line);margin-top:40px;padding-top:16px;color:var(--muted);font-size:12px}
@media(max-width:600px){.hero{flex-direction:column;text-align:center}.path{text-align:center}.stats{justify-content:center}}
`

func renderHTML(w io.Writer, p Payload) error {
	var b strings.Builder
	color := "#4ade80"
	switch {
	case p.Score < 50:
		color = "#f87171"
	case p.Score < 80:
		color = "#fbbf24"
	}

	findings := 0
	warns := 0
	for _, r := range p.Results {
		findings += len(r.Findings)
		if r.Status == "warn" || r.Error != "" {
			warns++
		}
	}

	b.WriteString("<!doctype html>\n<html lang=\"id\">\n<head>\n<meta charset=\"utf-8\">\n")
	b.WriteString("<meta name=\"viewport\" content=\"width=device-width,initial-scale=1\">\n")
	b.WriteString("<title>jaga — laporan keamanan</title>\n<style>")
	b.WriteString(htmlStyle)
	b.WriteString("</style>\n</head>\n<body>\n<main>\n")

	b.WriteString("<header><div class=\"brand\">jaga <span>🌿</span></div><div class=\"path\">")
	b.WriteString(html.EscapeString(p.Root))
	b.WriteString("</div></header>\n")

	b.WriteString("<section class=\"hero\"><div class=\"ring\" style=\"--score:")
	fmt.Fprintf(&b, "%d;--color:%s\"><div class=\"ring-inner\"><strong>%d</strong><span>/100</span></div></div>\n", p.Score, color, p.Score)
	b.WriteString("<div class=\"hero-copy\"><h1>Laporan keamanan repo</h1>")
	b.WriteString("<p>Hasil pemeriksaan otomatis. Semua proses berjalan lokal, tanpa telemetri.</p>")
	b.WriteString("<div class=\"stats\">")
	fmt.Fprintf(&b, "<div class=\"stat\"><b>%d</b><span>pemeriksaan</span></div>", len(p.Results))
	fmt.Fprintf(&b, "<div class=\"stat\"><b>%d</b><span>temuan</span></div>", findings)
	fmt.Fprintf(&b, "<div class=\"stat\"><b>%d</b><span>perlu perhatian</span></div>", warns)
	b.WriteString("</div></div></section>\n")

	b.WriteString("<section><h2>Pemeriksaan</h2>\n")
	for _, r := range p.Results {
		b.WriteString("<article class=\"card\"><div class=\"card-head\"><h3>")
		b.WriteString(html.EscapeString(r.Name))
		b.WriteString("</h3><span class=\"pill ")
		b.WriteString(html.EscapeString(string(r.Severity)))
		b.WriteString("\">")
		b.WriteString(html.EscapeString(string(r.Severity)))
		b.WriteString("</span><span class=\"status")
		switch {
		case r.Error != "":
			b.WriteString(" error\">error")
		case r.Status == "warn":
			b.WriteString(" warn\">peringatan")
		default:
			b.WriteString("\">lolos")
		}
		b.WriteString("</span></div>\n")

		if len(r.Findings) > 0 {
			b.WriteString("<ul class=\"findings\">\n")
			for _, f := range r.Findings {
				b.WriteString("<li><code>")
				loc := f.File
				if f.Line > 0 {
					loc += ":" + fmt.Sprint(f.Line)
				}
				b.WriteString(html.EscapeString(loc))
				b.WriteString("</code><span>")
				b.WriteString(html.EscapeString(f.Message))
				b.WriteString("</span></li>\n")
			}
			b.WriteString("</ul>\n")
		} else {
			b.WriteString("<p class=\"empty\">Tidak ada temuan.</p>\n")
		}
		if r.Error != "" {
			b.WriteString("<p class=\"empty\">")
			b.WriteString(html.EscapeString(r.Error))
			b.WriteString("</p>\n")
		}
		b.WriteString("</article>\n")
	}
	b.WriteString("</section>\n")

	b.WriteString("<footer>Dihasilkan oleh <strong>jaga</strong> — satu gerbang sebelum deploy.</footer>\n")
	b.WriteString("</main>\n</body>\n</html>\n")
	_, err := io.WriteString(w, b.String())
	return err
}

func renderMarkdown(w io.Writer, p Payload) error {
	var b strings.Builder
	statusEmoji := "🟢"
	switch {
	case p.Score < 50:
		statusEmoji = "🔴"
	case p.Score < 80:
		statusEmoji = "🟡"
	}

	totalFindings := 0
	for _, r := range p.Results {
		totalFindings += len(r.Findings)
	}

	fmt.Fprintf(&b, "### %s Jaga Report: %d/100\n\n", statusEmoji, p.Score)
	b.WriteString("| Status | Check | Severity | Findings |\n")
	b.WriteString("| :---: | :--- | :--- | :--- |\n")

	for _, r := range p.Results {
		mark := "✅"
		if r.Status == "warn" {
			mark = "⚠️"
		}
		if r.Error != "" {
			mark = "❌"
		}
		count := fmt.Sprintf("%d finding(s)", len(r.Findings))
		if len(r.Findings) == 0 {
			count = "clean"
		}
		fmt.Fprintf(&b, "| %s | `%s` | `%s` | %s |\n", mark, r.Name, r.Severity, count)
	}

	if totalFindings > 0 {
		fmt.Fprintf(&b, "\n<details>\n<summary><b>Findings Detail (%d)</b></summary>\n\n", totalFindings)
		for _, r := range p.Results {
			if len(r.Findings) == 0 {
				continue
			}
			fmt.Fprintf(&b, "**%s** (`%s`):\n", r.Name, r.Severity)
			for _, f := range r.Findings {
				if f.Line > 0 {
					fmt.Fprintf(&b, "- `%s:%d` — %s\n", f.File, f.Line, f.Message)
				} else {
					fmt.Fprintf(&b, "- `%s` — %s\n", f.File, f.Message)
				}
			}
			b.WriteString("\n")
		}
		b.WriteString("</details>\n")
	}

	b.WriteString("\n> Scanned with [jaga](https://github.com/Emil-02/jaga) — zero-dependency pre-commit & CI gate.\n")
	_, err := io.WriteString(w, b.String())
	return err
}
