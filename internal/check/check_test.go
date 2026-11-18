package check

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Emil-02/jaga/internal/config"
)

func defaultCfg() config.Config { return config.Default() }

func TestSecretRun(t *testing.T) {
	dir := t.TempDir()
	content := "aws := \"AKIA1234567890ABCDEF\"\nclean := \"hello\"\n"
	if err := os.WriteFile(filepath.Join(dir, "main.go"), []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}
	res := Secret{}.Run(context.Background(), dir, defaultCfg())
	if len(res.Findings) == 0 {
		t.Fatal("secret tidak terdeteksi")
	}
	if strings.Contains(res.Findings[0].Message, "AKIA1234567890ABCDEF") {
		t.Fatal("secret tidak di-mask")
	}
}

func TestPIIRun(t *testing.T) {
	dir := t.TempDir()
	content := "nik: 3204123456789012\nhp: 081234567890\n"
	if err := os.WriteFile(filepath.Join(dir, "data.txt"), []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}
	res := PII{}.Run(context.Background(), dir, defaultCfg())
	if len(res.Findings) < 2 {
		t.Fatalf("hanya %d temuan: %+v", len(res.Findings), res.Findings)
	}
	for _, f := range res.Findings {
		if strings.Contains(f.Message, "3204123456789012") || strings.Contains(f.Message, "081234567890") {
			t.Fatal("data pribadi tidak di-mask")
		}
	}
}

func TestHygieneRun(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "README.md"), []byte("hi"), 0o600); err != nil {
		t.Fatal(err)
	}
	res := Hygiene{}.Run(context.Background(), dir, defaultCfg())
	if len(res.Findings) != 3 {
		t.Fatalf("harusnya 3 kekurangan, dapat %d: %+v", len(res.Findings), res.Findings)
	}
}
