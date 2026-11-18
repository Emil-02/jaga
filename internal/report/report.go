package report

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"github.com/Emil-02/jaga/internal/check"
)

type Payload struct {
	Root    string         `json:"root"`
	Score   int            `json:"score"`
	Results []check.Result `json:"results"`
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
