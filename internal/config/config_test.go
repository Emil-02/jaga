package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoad(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, ".jaga.yml")
	content := `version: 1
exclude:
  - dist
  - vendor
checks:
  secret:
    enabled: true
    severity: critical
  pii:
    enabled: false
    severity: low
fail_on: medium
`
	if err := os.WriteFile(p, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}
	cfg, err := Load(p)
	if err != nil {
		t.Fatal(err)
	}
	if cfg.FailOn != SeverityMedium {
		t.Fatalf("FailOn = %s", cfg.FailOn)
	}
	if !cfg.For("secret").Enabled || cfg.For("secret").Severity != SeverityCritical {
		t.Fatalf("secret config = %+v", cfg.For("secret"))
	}
	if cfg.For("pii").Enabled {
		t.Fatal("pii harusnya nonaktif")
	}
	if !cfg.IsExcluded("dist") || cfg.IsExcluded("src") {
		t.Fatal("exclude tidak bekerja")
	}
}

func TestDefaultMissingFile(t *testing.T) {
	cfg, err := Load(filepath.Join(t.TempDir(), "tidak-ada.yml"))
	if err != nil {
		t.Fatal(err)
	}
	if cfg.FailOn != SeverityHigh || len(cfg.Checks) != 3 {
		t.Fatalf("default salah: %+v", cfg)
	}
}

func TestSeverityRank(t *testing.T) {
	if SeverityCritical.Rank() <= SeverityHigh.Rank() {
		t.Fatal("urutan severity salah")
	}
	if _, ok := ParseSeverity("HIGH"); !ok {
		t.Fatal("ParseSeverity gagal untuk huruf besar")
	}
}
