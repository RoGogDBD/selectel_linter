package loglint_test

import (
	"testing"

	"github.com/makarkudryavtsev/selectel-linter/internal/analyzer/loglint"
	"golang.org/x/tools/go/analysis/analysistest"
)

func TestAnalyzerCustomConfig(t *testing.T) {
	cfg := loglint.DefaultConfig()
	cfg.Rules.Lowercase = false
	cfg.Rules.SpecialSymbols = false
	cfg.SensitiveKeywords = []string{"token"}
	cfg.SensitivePatterns = []string{`(?i)ssn\s*[:=]`}

	an, err := loglint.NewAnalyzer(cfg)
	if err != nil {
		t.Fatalf("NewAnalyzer() error = %v", err)
	}

	analysistest.Run(t, analysistest.TestData(), an, "configured")
}
