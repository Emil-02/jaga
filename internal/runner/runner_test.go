package runner

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/Emil-02/jaga/internal/config"
)

func TestRunAll(t *testing.T) {
	dir := t.TempDir()
	for _, name := range []string{"README.md", "LICENSE"} {
		if err := os.WriteFile(filepath.Join(dir, name), []byte("x"), 0o600); err != nil {
			t.Fatal(err)
		}
	}
	if err := os.MkdirAll(filepath.Join(dir, ".github", "workflows"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, ".github", "workflows", "ci.yml"), []byte("x"), 0o600); err != nil {
		t.Fatal(err)
	}

	results := NewDefault().RunAll(context.Background(), dir, config.Default())
	if len(results) != 3 {
		t.Fatalf("dapat %d hasil: %+v", len(results), results)
	}
	for i := 1; i < len(results); i++ {
		if results[i-1].Name > results[i].Name {
			t.Fatal("hasil tidak terurut")
		}
	}

	cfg := config.Default()
	cc := cfg.Checks["secret"]
	cc.Enabled = false
	cfg.Checks["secret"] = cc
	if got := len(NewDefault().RunAll(context.Background(), dir, cfg)); got != 2 {
		t.Fatalf("check nonaktif tetap jalan: %d", got)
	}
}
