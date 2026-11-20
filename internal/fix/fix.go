package fix

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/Emil-02/jaga/internal/report"
)

const maxManualActions = 20

type Action struct {
	Check   string `json:"check"`
	Kind    string `json:"kind"`
	Title   string `json:"title"`
	Detail  string `json:"detail"`
	File    string `json:"file,omitempty"`
	Line    int    `json:"line,omitempty"`
	Path    string `json:"path,omitempty"`
	Content string `json:"content,omitempty"`
	Safe    bool   `json:"safe"`
}

type Plan struct {
	Root    string   `json:"root"`
	Score   int      `json:"score"`
	Actions []Action `json:"actions"`
	Applied []string `json:"applied,omitempty"`
}

func Build(root string, payload report.Payload, author string) Plan {
	plan := Plan{Root: root, Score: payload.Score, Actions: []Action{}}
	if strings.TrimSpace(author) == "" {
		author = "Emil-02"
	}

	for _, result := range payload.Results {
		switch result.Name {
		case "hygiene":
			for _, f := range result.Findings {
				switch {
				case strings.HasPrefix(f.Message, "LICENSE "):
					plan.Actions = append(plan.Actions, Action{
						Check: result.Name, Kind: "create", Path: "LICENSE", Safe: true,
						Title: "Tambah LICENSE MIT", Detail: "LICENSE tidak ditemukan di root repo.",
						Content: licenseContent(author),
					})
				case strings.HasPrefix(f.Message, "README "):
					plan.Actions = append(plan.Actions, Action{
						Check: result.Name, Kind: "create", Path: "README.md", Safe: true,
						Title: "Tambah README dasar", Detail: "README tidak ditemukan di root repo.",
						Content: readmeContent(filepath.Base(root)),
					})
				case strings.HasPrefix(f.Message, "CI workflow "):
					plan.Actions = append(plan.Actions, Action{
						Check: result.Name, Kind: "create", Path: ".github/workflows/jaga.yml", Safe: true,
						Title: "Tambah workflow CI jaga", Detail: "Workflow CI tidak ditemukan.",
						Content: workflowContent(),
					})
				case strings.HasPrefix(f.Message, ".gitignore "):
					plan.Actions = append(plan.Actions, Action{
						Check: result.Name, Kind: "append", Path: ".gitignore", Safe: true,
						Title: "Lindungi file rahasia lewat .gitignore", Detail: ".gitignore tidak ditemukan.",
						Content: ".env\n.env.*\n!.env.example\n*.pem\n*.key\n",
					})
				}
			}
		case "secret", "pii":
			for i, f := range result.Findings {
				if i >= maxManualActions {
					break
				}
				title, detail := manualAdvice(result.Name, f.Message)
				plan.Actions = append(plan.Actions, Action{
					Check: result.Name, Kind: "manual", Title: title, Detail: detail,
					File: f.File, Line: f.Line, Safe: false,
				})
			}
		}
	}
	return plan
}

func manualAdvice(check, message string) (string, string) {
	if check == "secret" {
		return "Rotasi kredensial", fmt.Sprintf(
			"%s. Jangan commit nilai asli — rotasi kunci, simpan di environment/secret manager, dan perbarui kode. Temuan: %s",
			message, message,
		)
	}
	return "Masking data pribadi", fmt.Sprintf(
		"Tutup atau hapus data pribadi ini sesuai UU PDP No. 27/2022, terutama bila tampil di README/log publik. Temuan: %s",
		message,
	)
}

func licenseContent(author string) string {
	return fmt.Sprintf(`MIT License

Copyright (c) %d %s

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
SOFTWARE.
`, time.Now().Year(), strings.TrimSpace(author))
}

func readmeContent(name string) string {
	return fmt.Sprintf(`# %s

## Tentang

Jelaskan secara singkat apa fungsi proyek ini.

## Menjalankan

`+"```sh\n"+"npm install\n"+"```"+`

## Lisensi

MIT.
`, strings.TrimSpace(name))
}

func workflowContent() string {
	return `name: jaga

on: [push, pull_request]

permissions:
  contents: read

jobs:
  jaga:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: Emil-02/jaga@main
        with:
          format: markdown
          summary: "true"
`
}

func Apply(root string, plan *Plan) ([]string, error) {
	applied := []string{}
	for _, action := range plan.Actions {
		if !action.Safe || (action.Kind != "create" && action.Kind != "append") {
			continue
		}
		path := filepath.Join(root, action.Path)
		rel, err := filepath.Rel(root, path)
		if err != nil || rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
			return applied, fmt.Errorf("path di luar root: %s", action.Path)
		}
		if action.Kind == "create" {
			if _, err := os.Stat(path); err == nil {
				continue
			}
			if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
				return applied, err
			}
			if err := os.WriteFile(path, []byte(action.Content), 0o644); err != nil {
				return applied, err
			}
			applied = append(applied, action.Path)
			continue
		}

		lines := []string{}
		if data, err := os.ReadFile(path); err == nil {
			lines = strings.Split(strings.TrimRight(string(data), "\n"), "\n")
		}
		seen := map[string]bool{}
		for _, line := range lines {
			seen[line] = true
		}
		changed := false
		for _, line := range strings.Split(strings.TrimRight(action.Content, "\n"), "\n") {
			if line != "" && !seen[line] {
				lines = append(lines, line)
				seen[line] = true
				changed = true
			}
		}
		if !changed {
			continue
		}
		if err := os.WriteFile(path, []byte(strings.Join(lines, "\n")+"\n"), 0o644); err != nil {
			return applied, err
		}
		applied = append(applied, action.Path)
	}
	plan.Applied = applied
	return applied, nil
}

func Render(w io.Writer, plan Plan, format string) error {
	switch format {
	case "json":
		return json.NewEncoder(w).Encode(plan)
	case "text":
		return renderText(w, plan)
	case "patch":
		return renderPatch(w, plan)
	default:
		return fmt.Errorf("format tidak dikenal: %s", format)
	}
}

func renderText(w io.Writer, plan Plan) error {
	var b strings.Builder
	fmt.Fprintf(&b, "jaga fix — %s (skor %d)\n", plan.Root, plan.Score)
	fmt.Fprintf(&b, "%d perbaikan aman, %d perlu tindakan manual\n\n", countSafe(plan), len(plan.Actions)-countSafe(plan))
	for i, action := range plan.Actions {
		mark := "manual"
		if action.Safe {
			mark = "aman"
		}
		fmt.Fprintf(&b, "%2d. [%s] %s\n", i+1, mark, action.Title)
		if action.Path != "" {
			fmt.Fprintf(&b, "    file: %s\n", action.Path)
		} else {
			fmt.Fprintf(&b, "    file: %s:%d\n", action.File, action.Line)
		}
		fmt.Fprintf(&b, "    %s\n", action.Detail)
	}
	_, err := io.WriteString(w, b.String())
	return err
}

func countSafe(plan Plan) int {
	n := 0
	for _, a := range plan.Actions {
		if a.Safe {
			n++
		}
	}
	return n
}

func renderPatch(w io.Writer, plan Plan) error {
	var b strings.Builder
	for _, action := range plan.Actions {
		if !action.Safe {
			continue
		}
		switch action.Kind {
		case "create":
			lines := strings.Split(strings.TrimRight(action.Content, "\n"), "\n")
			fmt.Fprintf(&b, "--- /dev/null\n+++ b/%s\n@@ -0,0 +1,%d @@\n", action.Path, len(lines))
			for _, line := range lines {
				b.WriteString("+" + line + "\n")
			}
		case "append":
			fmt.Fprintf(&b, "--- a/%s\n+++ b/%s\n", action.Path, action.Path)
			for _, line := range strings.Split(strings.TrimRight(action.Content, "\n"), "\n") {
				if line != "" {
					b.WriteString("+" + line + "\n")
				}
			}
		}
		b.WriteString("\n")
	}
	_, err := io.WriteString(w, b.String())
	return err
}
