package main

import (
	"github.com/makarkudryavtsev/selectel-linter/internal/analyzer/loglint"
	"golang.org/x/tools/go/analysis/multichecker"
)

func main() {
	multichecker.Main(loglint.Analyzer)
}
