package config

import (
	"bufio"
	"errors"
	"os"
	"path/filepath"
	"strings"
)

type Severity string

const (
	SeverityLow      Severity = "low"
	SeverityMedium   Severity = "medium"
	SeverityHigh     Severity = "high"
	SeverityCritical Severity = "critical"
)

var severityRank = map[Severity]int{
	SeverityLow: 1, SeverityMedium: 2, SeverityHigh: 3, SeverityCritical: 4,
}

func ParseSeverity(s string) (Severity, bool) {
	v := Severity(strings.ToLower(strings.TrimSpace(s)))
	_, ok := severityRank[v]
	return v, ok
}

func (s Severity) Rank() int {
	return severityRank[s]
}

type CheckConfig struct {
	Enabled  bool
	Severity Severity
}

type Config struct {
	Version int
	Exclude []string
	Checks  map[string]CheckConfig
	FailOn  Severity
}

func Default() Config {
	return Config{
		Version: 1,
		Exclude: []string{".git", "node_modules", "dist", "vendor", "build", "coverage", "testdata"},
		Checks: map[string]CheckConfig{
			"secret":  {Enabled: true, Severity: SeverityCritical},
			"pii":     {Enabled: true, Severity: SeverityHigh},
			"hygiene": {Enabled: true, Severity: SeverityLow},
		},
		FailOn: SeverityHigh,
	}
}

func (c Config) IsExcluded(name string) bool {
	for _, ex := range c.Exclude {
		if name == ex {
			return true
		}
	}
	return false
}

func (c Config) For(name string) CheckConfig {
	if v, ok := c.Checks[name]; ok {
		return v
	}
	return CheckConfig{Enabled: true, Severity: SeverityMedium}
}

func Load(path string) (Config, error) {
	cfg := Default()
	if path == "" {
		return cfg, nil
	}
	if _, err := os.Stat(path); errors.Is(err, os.ErrNotExist) {
		return cfg, nil
	} else if err != nil {
		return cfg, err
	}

	f, err := os.Open(path)
	if err != nil {
		return cfg, err
	}
	defer f.Close()

	section := ""
	currentCheck := ""
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		raw := strings.TrimRight(scanner.Text(), "\r")
		trimmed := strings.TrimSpace(raw)
		if trimmed == "" || strings.HasPrefix(trimmed, "#") {
			continue
		}
		indent := len(raw) - len(strings.TrimLeft(raw, " \t"))
		key, value, hasValue := splitKV(trimmed)

		switch {
		case indent == 0 && hasValue:
			currentCheck = ""
			switch key {
			case "version":
				cfg.Version = atoi(value)
			case "exclude":
				section = "exclude"
			case "checks":
				section = "checks"
			case "fail_on":
				section = ""
				if sev, ok := ParseSeverity(strings.Trim(value, "\"'")); ok {
					cfg.FailOn = sev
				}
			}
		case section == "exclude" && strings.HasPrefix(trimmed, "-"):
			item := strings.TrimSpace(strings.TrimPrefix(trimmed, "-"))
			item = strings.Trim(item, "\"'")
			if item != "" {
				cfg.Exclude = append(cfg.Exclude, item)
			}
		case section == "checks" && indent == 2 && value == "":
			currentCheck = key
		case section == "checks" && indent >= 4 && currentCheck != "" && hasValue:
			cc := cfg.Checks[currentCheck]
			switch key {
			case "enabled":
				cc.Enabled = strings.EqualFold(value, "true")
			case "severity":
				if sev, ok := ParseSeverity(strings.Trim(value, "\"'")); ok {
					cc.Severity = sev
				}
			}
			cfg.Checks[currentCheck] = cc
		}
	}
	if err := scanner.Err(); err != nil {
		return cfg, err
	}
	return cfg, nil
}

func splitKV(s string) (string, string, bool) {
	if i := strings.Index(s, ":"); i >= 0 {
		return strings.TrimSpace(s[:i]), strings.TrimSpace(s[i+1:]), true
	}
	return s, "", false
}

func atoi(s string) int {
	n := 0
	for _, r := range strings.Trim(s, "\"'") {
		if r < '0' || r > '9' {
			return 0
		}
		n = n*10 + int(r-'0')
	}
	return n
}

func Find(root string) string {
	for {
		p := filepath.Join(root, ".jaga.yml")
		if _, err := os.Stat(p); err == nil {
			return p
		}
		parent := filepath.Dir(root)
		if parent == root {
			return ""
		}
		root = parent
	}
}
