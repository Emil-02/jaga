package report

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"

	"github.com/Emil-02/jaga/internal/check"
)

func sample() Payload {
	return Payload{
		Root:  "/tmp/demo",
		Score: 88,
		Results: []check.Result{
			{Name: "secret", Status: "pass", Severity: "critical", Findings: []check.Finding{}},
			{Name: "pii", Status: "warn", Severity: "high", Findings: []check.Finding{{File: "README.md", Line: 3, Message: "NIK: 3204…9012"}}},
		},
	}
}

func TestRenderAllFormats(t *testing.T) {
	for _, format := range []string{"text", "json", "sarif", "badge"} {
		var buf bytes.Buffer
		if err := Render(&buf, sample(), format); err != nil {
			t.Fatalf("%s: %v", format, err)
		}
		if buf.Len() == 0 {
			t.Fatalf("%s: keluaran kosong", format)
		}
		if format == "json" || format == "sarif" {
			var v any
			if err := json.Unmarshal(buf.Bytes(), &v); err != nil {
				t.Fatalf("%s bukan JSON valid: %v", format, err)
			}
		}
		if format == "badge" && !strings.Contains(buf.String(), "<svg") {
			t.Fatal("badge bukan SVG")
		}
	}
}

func TestRenderUnknownFormat(t *testing.T) {
	var buf bytes.Buffer
	if err := Render(&buf, sample(), "xml"); err == nil {
		t.Fatal("format asing harusnya error")
	}
}
