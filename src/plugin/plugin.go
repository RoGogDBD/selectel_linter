package main

import (
	"github.com/makarkudryavtsev/selectel-linter/internal/analyzer/loglint"
	"golang.org/x/tools/go/analysis"
)

func New(conf any) ([]*analysis.Analyzer, error) {
	return loglint.New(conf)
}
