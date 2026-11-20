package fix

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Emil-02/jaga/internal/check"
	"github.com/Emil-02/jaga/internal/config"
	"github.com/Emil-02/jaga/internal/report"
)

func samplePayload() report.Payload {
	return report.Payload{
		Score: 60,
		Results: []check.Result{
			{
				Name: "hygiene", Status: "warn", Severity: config.SeverityLow,
				Findings: []check.Finding{
					{File: ".", Message: "LICENSE tidak ditemukan", Severity: config.SeverityLow},
					{File: ".", Message: "README tidak ditemukan", Severity: config.SeverityLow},
					{File: ".", Message: "CI workflow tidak ditemukan", Severity: config.SeverityLow},
					{File: ".", Message: ".gitignore tidak ditemukan", Severity: config.SeverityLow},
				},
			},
			{
				Name: "secret", Status: "warn", Severity: config.SeverityCritical,
				Findings: []check.Finding{{File: "app.js", Line: 3, Message: "OpenAI API key: sk-a…7890", Severity: config.SeverityCritical}},
			},
			{
				Name: "pii", Status: "warn", Severity: config.SeverityHigh,
				Findings: []check.Finding{{File: "README.md", Line: 9, Message: "NIK/KK: 3204…9012", Severity: config.SeverityHigh}},
			},
		},
	}
}

func TestBuildPlan(t *testing.T) {
	dir := t.TempDir()
	plan := Build(dir, samplePayload(), "emil rahmat")
	if len(plan.Actions) != 6 {
		t.Fatalf("actions = %d", len(plan.Actions))
	}
	safe, manual := 0, 0
	for _, a := range plan.Actions {
		if a.Safe {
			safe++
			if a.Kind == "create" && a.Path == "LICENSE" && !strings.Contains(a.Content, "emil rahmat") {
				t.Fatal("LICENSE tidak memakai author")
			}
		} else {
			manual++
			if a.Kind != "manual" || a.File == "" {
				t.Fatalf("manual action tidak valid: %+v", a)
			}
		}
	}
	if safe != 4 || manual != 2 {
		t.Fatalf("safe=%d manual=%d", safe, manual)
	}
}

func TestApplyCreatesAndIsIdempotent(t *testing.T) {
	dir := t.TempDir()
	plan := Build(dir, samplePayload(), "emil rahmat")
	applied, err := Apply(dir, &plan)
	if err != nil {
		t.Fatal(err)
	}
	if len(applied) != 4 {
		t.Fatalf("applied = %v", applied)
	}
	for _, name := range []string{"LICENSE", "README.md", ".github/workflows/jaga.yml", ".gitignore"} {
		if _, err := os.Stat(filepath.Join(dir, name)); err != nil {
			t.Fatalf("%s tidak dibuat: %v", name, err)
		}
	}

	second := Build(dir, samplePayload(), "emil rahmat")
	applied, err = Apply(dir, &second)
	if err != nil {
		t.Fatal(err)
	}
	if len(applied) != 0 {
		t.Fatalf("apply kedua menulis ulang: %v", applied)
	}
}

func TestApplyDoesNotOverwriteAndDeduplicates(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "LICENSE"), []byte("CUSTOM\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, ".gitignore"), []byte(".env\nnode_modules\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	plan := Build(dir, samplePayload(), "emil rahmat")
	if _, err := Apply(dir, &plan); err != nil {
		t.Fatal(err)
	}
	if got := mustRead(t, filepath.Join(dir, "LICENSE")); got != "CUSTOM\n" {
		t.Fatalf("LICENSE ditimpa: %q", got)
	}
	gitignore := mustRead(t, filepath.Join(dir, ".gitignore"))
	if strings.Count(gitignore, ".env\n") != 1 || !strings.Contains(gitignore, "node_modules") || !strings.Contains(gitignore, "*.pem") {
		t.Fatalf(".gitignore = %q", gitignore)
	}
}

func TestApplyRejectsTraversal(t *testing.T) {
	dir := t.TempDir()
	plan := Plan{Actions: []Action{{Kind: "create", Safe: true, Path: "../luar.txt", Content: "x"}}}
	if _, err := Apply(dir, &plan); err == nil {
		t.Fatal("path traversal harus ditolak")
	}
}

func TestRenderFormats(t *testing.T) {
	plan := Build(t.TempDir(), samplePayload(), "emil rahmat")
	var jsonBuf, patchBuf, textBuf bytes.Buffer
	if err := Render(&jsonBuf, plan, "json"); err != nil {
		t.Fatal(err)
	}
	var decoded Plan
	if err := json.Unmarshal(jsonBuf.Bytes(), &decoded); err != nil {
		t.Fatal(err)
	}
	if err := Render(&patchBuf, plan, "patch"); err != nil {
		t.Fatal(err)
	}
	if err := Render(&textBuf, plan, "text"); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(patchBuf.String(), "+++ b/LICENSE") {
		t.Fatal("patch tidak mengandung LICENSE")
	}
	if !strings.Contains(textBuf.String(), "perlu tindakan manual") {
		t.Fatal("text tidak merangkum manual")
	}
}

func mustRead(t *testing.T, path string) string {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return string(data)
}
