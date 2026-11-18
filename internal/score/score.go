package score

import (
	"github.com/Emil-02/jaga/internal/check"
	"github.com/Emil-02/jaga/internal/config"
)

var penalty = map[config.Severity]int{
	config.SeverityLow:      2,
	config.SeverityMedium:   5,
	config.SeverityHigh:     12,
	config.SeverityCritical: 25,
}

func Calculate(results []check.Result) int {
	total := 100
	for _, r := range results {
		for range r.Findings {
			total -= penalty[r.Severity]
		}
		if r.Error != "" {
			total -= 10
		}
	}
	if total < 0 {
		return 0
	}
	return total
}

func Breached(results []check.Result, failOn config.Severity) bool {
	for _, r := range results {
		if r.Severity.Rank() >= failOn.Rank() && len(r.Findings) > 0 {
			return true
		}
	}
	return false
}
