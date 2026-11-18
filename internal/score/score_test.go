package score

import (
	"testing"

	"github.com/Emil-02/jaga/internal/check"
	"github.com/Emil-02/jaga/internal/config"
)

func TestCalculateAndBreach(t *testing.T) {
	results := []check.Result{
		{Name: "secret", Severity: config.SeverityCritical, Status: "warn", Findings: []check.Finding{{Message: "x"}}},
		{Name: "hygiene", Severity: config.SeverityLow, Status: "warn", Findings: []check.Finding{{Message: "y"}}},
	}
	if got := Calculate(results); got != 73 {
		t.Fatalf("skor = %d, ingin 73", got)
	}
	if !Breached(results, config.SeverityHigh) {
		t.Fatal("harusnya breach pada high")
	}
	if Breached(results, config.SeverityCritical) == false {
		t.Fatal("critical juga harus breach")
	}
	clean := []check.Result{{Name: "secret", Severity: config.SeverityCritical, Findings: []check.Finding{}}}
	if Breached(clean, config.SeverityHigh) {
		t.Fatal("tidak boleh breach tanpa temuan")
	}
	if got := Calculate(clean); got != 100 {
		t.Fatalf("skor bersih = %d", got)
	}
}
