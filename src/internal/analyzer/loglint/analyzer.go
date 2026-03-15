package loglint

import (
	"fmt"
	"go/ast"
	"go/constant"
	"go/token"
	"go/types"
	"regexp"
	"strconv"
	"strings"
	"unicode"
	"unicode/utf8"

	"golang.org/x/tools/go/analysis"
)

var (
	defaultSensitiveKeywords = []string{
		"password", "passwd", "pwd", "token", "api_key", "apikey", "secret", "authorization", "bearer", "access_token", "refresh_token", "client_secret",
	}

	logMethods = map[string]struct{}{
		"Debug": {}, "Info": {}, "Warn": {}, "Warning": {}, "Error": {},
		"DPanic": {}, "Panic": {}, "Fatal": {},
		"Debugf": {}, "Infof": {}, "Warnf": {}, "Errorf": {}, "Panicf": {}, "Fatalf": {},
		"Debugw": {}, "Infow": {}, "Warnw": {}, "Errorw": {},
		"DebugContext": {}, "InfoContext": {}, "WarnContext": {}, "ErrorContext": {},
		"Log": {}, "LogAttrs": {},
	}
)

// RulesConfig configures which rules to apply.
type RulesConfig struct {
	Lowercase      bool
	English        bool
	SpecialSymbols bool
	SensitiveData  bool
}

// Config configures analyzer behavior.
type Config struct {
	Rules             RulesConfig
	SensitiveKeywords []string
	SensitivePatterns []string
}

type runtimeConfig struct {
	rules             RulesConfig
	sensitiveKeywords []string
	sensitivePatterns []*regexp.Regexp
}

// Analyzer validates log messages for slog and zap with default settings.
var Analyzer = mustAnalyzer(DefaultConfig())

// DefaultConfig returns analyzer defaults.
func DefaultConfig() Config {
	return Config{
		Rules: RulesConfig{
			Lowercase:      true,
			English:        true,
			SpecialSymbols: true,
			SensitiveData:  true,
		},
		SensitiveKeywords: append([]string(nil), defaultSensitiveKeywords...),
	}
}

func mustAnalyzer(cfg Config) *analysis.Analyzer {
	a, err := NewAnalyzer(cfg)
	if err != nil {
		panic(err)
	}
	return a
}

// NewAnalyzer creates an analyzer with custom settings.
func NewAnalyzer(cfg Config) (*analysis.Analyzer, error) {
	rc, err := buildRuntimeConfig(cfg)
	if err != nil {
		return nil, err
	}

	return &analysis.Analyzer{
		Name: "loglint",
		Doc:  "checks log messages for style, language, symbols and potential sensitive data leakage",
		Run: func(pass *analysis.Pass) (any, error) {
			return run(pass, rc)
		},
	}, nil
}

func buildRuntimeConfig(cfg Config) (runtimeConfig, error) {
	rules := cfg.Rules
	// Default to enabled for all rules if config omitted fields.
	if rules == (RulesConfig{}) {
		rules = DefaultConfig().Rules
	}

	keywords := cfg.SensitiveKeywords
	if len(keywords) == 0 {
		keywords = append([]string(nil), defaultSensitiveKeywords...)
	}
	for i := range keywords {
		keywords[i] = strings.ToLower(strings.TrimSpace(keywords[i]))
	}

	patterns := make([]*regexp.Regexp, 0, len(cfg.SensitivePatterns))
	for _, raw := range cfg.SensitivePatterns {
		p := strings.TrimSpace(raw)
		if p == "" {
			continue
		}
		re, err := regexp.Compile(p)
		if err != nil {
			return runtimeConfig{}, fmt.Errorf("compile sensitive pattern %q: %w", p, err)
		}
		patterns = append(patterns, re)
	}

	return runtimeConfig{
		rules:             rules,
		sensitiveKeywords: keywords,
		sensitivePatterns: patterns,
	}, nil
}

func run(pass *analysis.Pass, cfg runtimeConfig) (any, error) {
	for _, file := range pass.Files {
		ast.Inspect(file, func(n ast.Node) bool {
			call, ok := n.(*ast.CallExpr)
			if !ok {
				return true
			}

			method, msgExpr, extraArgs, ok := parseLogCall(pass, call)
			if !ok {
				return true
			}

			message, isConstString := constString(pass, msgExpr)
			if isConstString {
				if cfg.rules.Lowercase {
					reportLowercase(pass, msgExpr, message)
				}
				if cfg.rules.English {
					checkEnglish(pass, msgExpr.Pos(), message)
				}
				if cfg.rules.SpecialSymbols && !isFormatMethod(method) {
					checkSpecialSymbols(pass, msgExpr.Pos(), message)
				}
				if cfg.rules.SensitiveData {
					checkLiteralSensitive(pass, msgExpr.Pos(), message, cfg)
				}
			}
			if cfg.rules.SensitiveData {
				checkSensitiveExpression(pass, msgExpr, cfg)
				checkSensitiveArgs(pass, extraArgs, cfg)
			}

			return true
		})
	}

	return nil, nil
}

func parseLogCall(pass *analysis.Pass, call *ast.CallExpr) (string, ast.Expr, []ast.Expr, bool) {
	if len(call.Args) == 0 {
		return "", nil, nil, false
	}

	sel, ok := call.Fun.(*ast.SelectorExpr)
	if !ok {
		return "", nil, nil, false
	}

	if _, ok := logMethods[sel.Sel.Name]; !ok {
		return "", nil, nil, false
	}

	if !isSlogCall(pass, sel) && !isZapCall(pass, sel) {
		return "", nil, nil, false
	}

	msgIndex, ok := messageArgIndex(sel.Sel.Name, len(call.Args))
	if !ok {
		return "", nil, nil, false
	}

	extraArgs := []ast.Expr{}
	if msgIndex+1 < len(call.Args) {
		extraArgs = call.Args[msgIndex+1:]
	}

	return sel.Sel.Name, call.Args[msgIndex], extraArgs, true
}

func messageArgIndex(method string, argsCount int) (int, bool) {
	switch method {
	case "DebugContext", "InfoContext", "WarnContext", "ErrorContext":
		if argsCount < 2 {
			return 0, false
		}
		return 1, true
	case "Log", "LogAttrs":
		if argsCount < 3 {
			return 0, false
		}
		return 2, true
	default:
		return 0, true
	}
}

func isFormatMethod(method string) bool {
	return strings.HasSuffix(method, "f")
}

func isSlogCall(pass *analysis.Pass, sel *ast.SelectorExpr) bool {
	if ident, ok := sel.X.(*ast.Ident); ok {
		if pkgName, ok := pass.TypesInfo.Uses[ident].(*types.PkgName); ok {
			if pkgName.Imported().Path() == "log/slog" {
				return true
			}
		}
	}

	selection := pass.TypesInfo.Selections[sel]
	if selection == nil {
		return false
	}
	return typeFromPackage(selection.Recv(), "log/slog", "Logger")
}

func isZapCall(pass *analysis.Pass, sel *ast.SelectorExpr) bool {
	selection := pass.TypesInfo.Selections[sel]
	if selection == nil {
		return false
	}

	recv := selection.Recv()
	return typeFromPackage(recv, "go.uber.org/zap", "Logger") ||
		typeFromPackage(recv, "go.uber.org/zap", "SugaredLogger")
}

func typeFromPackage(t types.Type, pkgPath, typeName string) bool {
	if ptr, ok := t.(*types.Pointer); ok {
		t = ptr.Elem()
	}
	named, ok := t.(*types.Named)
	if !ok {
		return false
	}
	obj := named.Obj()
	if obj == nil || obj.Pkg() == nil {
		return false
	}
	return obj.Pkg().Path() == pkgPath && obj.Name() == typeName
}

func constString(pass *analysis.Pass, expr ast.Expr) (string, bool) {
	tv, ok := pass.TypesInfo.Types[expr]
	if !ok || tv.Value == nil || tv.Value.Kind() != constant.String {
		return "", false
	}
	return constant.StringVal(tv.Value), true
}

func reportLowercase(pass *analysis.Pass, expr ast.Expr, msg string) {
	trimmed := strings.TrimSpace(msg)
	if trimmed == "" {
		return
	}

	for _, r := range trimmed {
		if !unicode.IsLetter(r) {
			continue
		}
		if unicode.IsLower(r) {
			return
		}
		break
	}

	d := analysis.Diagnostic{
		Pos:     expr.Pos(),
		Message: "log message must start with lowercase letter",
	}

	if lit, ok := expr.(*ast.BasicLit); ok && lit.Kind == token.STRING {
		fixed, ok := lowercaseFirstLetter(msg)
		if ok {
			d.SuggestedFixes = []analysis.SuggestedFix{{
				Message: "convert first letter to lowercase",
				TextEdits: []analysis.TextEdit{{
					Pos:     lit.Pos(),
					End:     lit.End(),
					NewText: []byte(strconv.Quote(fixed)),
				}},
			}}
		}
	}

	pass.Report(d)
}

func lowercaseFirstLetter(s string) (string, bool) {
	for i, r := range s {
		if !unicode.IsLetter(r) {
			continue
		}
		if unicode.IsLower(r) {
			return "", false
		}
		rn := unicode.ToLower(r)
		size := utf8.RuneLen(r)
		if size < 0 {
			return "", false
		}
		return s[:i] + string(rn) + s[i+size:], true
	}
	return "", false
}

func checkEnglish(pass *analysis.Pass, pos token.Pos, msg string) {
	for _, r := range msg {
		if !unicode.IsLetter(r) {
			continue
		}
		if r > unicode.MaxASCII {
			pass.Reportf(pos, "log message must contain only english text")
			return
		}
	}
}

func checkSpecialSymbols(pass *analysis.Pass, pos token.Pos, msg string) {
	for _, r := range msg {
		switch {
		case unicode.IsLetter(r):
		case unicode.IsDigit(r):
		case r == ' ':
		default:
			pass.Reportf(pos, "log message must not contain special symbols or emoji")
			return
		}
	}
}

func checkLiteralSensitive(pass *analysis.Pass, pos token.Pos, msg string, cfg runtimeConfig) {
	if containsSensitivePrefixOrPattern(msg, cfg) {
		pass.Reportf(pos, "log message may contain sensitive data")
	}
}

func checkSensitiveExpression(pass *analysis.Pass, expr ast.Expr, cfg runtimeConfig) {
	if exprContainsDynamicSensitiveData(pass, expr, cfg) {
		pass.Reportf(expr.Pos(), "log message may contain sensitive data")
	}
}

func exprContainsDynamicSensitiveData(pass *analysis.Pass, expr ast.Expr, cfg runtimeConfig) bool {
	if exprContainsDynamicSensitiveConcatenation(pass, expr, cfg) {
		return true
	}
	call, ok := expr.(*ast.CallExpr)
	if !ok {
		return false
	}
	return callContainsSensitiveFormatting(pass, call, cfg)
}

func exprContainsDynamicSensitiveConcatenation(pass *analysis.Pass, expr ast.Expr, cfg runtimeConfig) bool {
	bin, ok := expr.(*ast.BinaryExpr)
	if !ok || bin.Op != token.ADD {
		return false
	}

	parts := flattenConcat(bin)
	hasSensitiveLiteral := false
	hasNonConstPart := false

	for _, part := range parts {
		if s, ok := constString(pass, part); ok {
			if containsSensitivePrefixOrPattern(s, cfg) {
				hasSensitiveLiteral = true
			}
			continue
		}
		hasNonConstPart = true
	}

	return hasSensitiveLiteral && hasNonConstPart
}

func flattenConcat(expr ast.Expr) []ast.Expr {
	bin, ok := expr.(*ast.BinaryExpr)
	if !ok || bin.Op != token.ADD {
		return []ast.Expr{expr}
	}
	return append(flattenConcat(bin.X), flattenConcat(bin.Y)...)
}

func callContainsSensitiveFormatting(pass *analysis.Pass, call *ast.CallExpr, cfg runtimeConfig) bool {
	if len(call.Args) < 2 || !isFmtSprintfCall(pass, call.Fun) {
		return false
	}

	format, ok := constString(pass, call.Args[0])
	if !ok || !containsSensitivePrefixOrPattern(format, cfg) {
		return false
	}

	for _, arg := range call.Args[1:] {
		if _, ok := constString(pass, arg); !ok {
			return true
		}
	}
	return false
}

func isFmtSprintfCall(pass *analysis.Pass, fun ast.Expr) bool {
	sel, ok := fun.(*ast.SelectorExpr)
	if !ok || sel.Sel.Name != "Sprintf" {
		return false
	}
	ident, ok := sel.X.(*ast.Ident)
	if !ok {
		return false
	}
	pkgName, ok := pass.TypesInfo.Uses[ident].(*types.PkgName)
	if !ok {
		return false
	}
	return pkgName.Imported().Path() == "fmt"
}

func checkSensitiveArgs(pass *analysis.Pass, args []ast.Expr, cfg runtimeConfig) {
	for _, arg := range args {
		if containsSensitiveArg(pass, arg, cfg) {
			pass.Reportf(arg.Pos(), "log message may contain sensitive data")
			return
		}
	}
}

func containsSensitiveArg(pass *analysis.Pass, arg ast.Expr, cfg runtimeConfig) bool {
	if s, ok := constString(pass, arg); ok {
		return containsSensitiveKeyword(s, cfg)
	}

	call, ok := arg.(*ast.CallExpr)
	if !ok || len(call.Args) == 0 {
		return false
	}
	first, ok := constString(pass, call.Args[0])
	if !ok {
		return false
	}
	return containsSensitiveKeyword(first, cfg)
}

func containsSensitivePrefixOrPattern(s string, cfg runtimeConfig) bool {
	if containsSensitivePrefix(s, cfg) {
		return true
	}
	for _, re := range cfg.sensitivePatterns {
		if re.MatchString(s) {
			return true
		}
	}
	return false
}

func containsSensitivePrefix(s string, cfg runtimeConfig) bool {
	lower := strings.ToLower(s)
	for _, word := range cfg.sensitiveKeywords {
		if word == "" {
			continue
		}
		if strings.Contains(lower, word+":") || strings.Contains(lower, word+"=") {
			return true
		}
	}
	return false
}

func containsSensitiveKeyword(s string, cfg runtimeConfig) bool {
	lower := strings.ToLower(s)
	for _, word := range cfg.sensitiveKeywords {
		if word != "" && strings.Contains(lower, word) {
			return true
		}
	}
	return false
}

// New is golangci-lint go-plugin entrypoint.
func New(conf any) ([]*analysis.Analyzer, error) {
	cfg, err := ParseConfig(conf)
	if err != nil {
		return nil, err
	}
	a, err := NewAnalyzer(cfg)
	if err != nil {
		return nil, err
	}
	return []*analysis.Analyzer{a}, nil
}
