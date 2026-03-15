package moduleplugin

import (
	"github.com/golangci/plugin-module-register/register"
	"github.com/makarkudryavtsev/selectel-linter/internal/analyzer/loglint"
	"golang.org/x/tools/go/analysis"
)

type plugin struct {
	analyzer *analysis.Analyzer
}

func init() {
	register.Plugin("loglint", New)
}

// New — точка входа для режима module plugin в golangci-lint.
func New(conf any) (register.LinterPlugin, error) {
	cfg, err := loglint.ParseConfig(conf)
	if err != nil {
		return nil, err
	}
	a, err := loglint.NewAnalyzer(cfg)
	if err != nil {
		return nil, err
	}
	return &plugin{analyzer: a}, nil
}

func (p *plugin) BuildAnalyzers() ([]*analysis.Analyzer, error) {
	return []*analysis.Analyzer{p.analyzer}, nil
}

func (p *plugin) GetLoadMode() string {
	return register.LoadModeTypesInfo
}
