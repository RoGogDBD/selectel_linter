package loglint_test

import (
	"testing"

	"github.com/makarkudryavtsev/selectel-linter/internal/analyzer/loglint"
	"golang.org/x/tools/go/analysis/analysistest"
)

func TestAnalyzer(t *testing.T) {
	testData := analysistest.TestData()
	cases := []string{
		"a",
		"contexts",
		"lowercase",
		"english",
		"symbols",
		"sensitive",
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc, func(t *testing.T) {
			analysistest.Run(t, testData, loglint.Analyzer, tc)
		})
	}
}
