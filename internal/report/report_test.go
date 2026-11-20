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
	for _, format := range []string{"text", "json", "sarif", "badge", "html"} {
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

func TestRenderHTMLEscapesFindings(t *testing.T) {
	p := sample()
	p.Root = `/tmp/<repo> & "saudara"`
	p.Results[1].Findings[0].File = `<img src=x onerror=alert(1)>`
	p.Results[1].Findings[0].Message = `NIK & <data> "rahasia"`
	var buf bytes.Buffer
	if err := Render(&buf, p, "html"); err != nil {
		t.Fatal(err)
	}
	out := buf.String()
	for _, want := range []string{"<!doctype html>", "--score:88", "&lt;img", "&lt;data&gt;", "&amp;", "&#34;rahasia&#34;"} {
		if !strings.Contains(out, want) {
			t.Fatalf("HTML tidak mengandung %q", want)
		}
	}
	if strings.Contains(out, "<img src=x") {
		t.Fatal("HTML tidak di-escape")
	}
}

func TestRenderUnknownFormat(t *testing.T) {
	var buf bytes.Buffer
	if err := Render(&buf, sample(), "xml"); err == nil {
		t.Fatal("format asing harusnya error")
	}
}
