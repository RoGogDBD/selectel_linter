package loglint

import (
	"fmt"
	"go/ast"
	"go/constant"
	"go/token"
	"go/types"
	"strings"
	"unicode"

	"golang.org/x/tools/go/analysis"
)

var (
	// Analyzer validates log messages for slog and zap.
	Analyzer = &analysis.Analyzer{
		Name: "loglint",
		Doc:  "checks log messages for style, language, symbols and potential sensitive data leakage",
		Run:  run,
	}

	logMethods = map[string]struct{}{
		"Debug": {}, "Info": {}, "Warn": {}, "Warning": {}, "Error": {},
		"DPanic": {}, "Panic": {}, "Fatal": {},
		"Debugf": {}, "Infof": {}, "Warnf": {}, "Errorf": {}, "Panicf": {}, "Fatalf": {},
		"Debugw": {}, "Infow": {}, "Warnw": {}, "Errorw": {},
		"DebugContext": {}, "InfoContext": {}, "WarnContext": {}, "ErrorContext": {},
		"Log": {}, "LogAttrs": {},
	}

	sensitiveWords = []string{
		"password", "passwd", "pwd", "token", "api_key", "apikey", "secret", "auth", "authorization", "bearer",
	}
)

func run(pass *analysis.Pass) (any, error) {
	for _, file := range pass.Files {
		ast.Inspect(file, func(n ast.Node) bool {
			call, ok := n.(*ast.CallExpr)
			if !ok {
				return true
			}

			msgExpr, ok := extractLogMessageExpr(pass, call)
			if !ok {
				return true
			}

			message, isConstString := constString(pass, msgExpr)
			if isConstString {
				checkLowercase(pass, msgExpr.Pos(), message)
				checkEnglish(pass, msgExpr.Pos(), message)
				checkSpecialSymbols(pass, msgExpr.Pos(), message)
				checkLiteralSensitive(pass, msgExpr.Pos(), message)
			}
			checkSensitiveExpression(pass, msgExpr)

			return true
		})
	}

	return nil, nil
}

func extractLogMessageExpr(pass *analysis.Pass, call *ast.CallExpr) (ast.Expr, bool) {
	if len(call.Args) == 0 {
		return nil, false
	}

	sel, ok := call.Fun.(*ast.SelectorExpr)
	if !ok {
		return nil, false
	}

	if _, ok := logMethods[sel.Sel.Name]; !ok {
		return nil, false
	}

	if !isSlogCall(pass, sel) && !isZapCall(pass, sel) {
		return nil, false
	}

	firstArg, ok := messageArg(sel.Sel.Name, call.Args)
	if !ok {
		return nil, false
	}

	return firstArg, true
}

func messageArg(method string, args []ast.Expr) (ast.Expr, bool) {
	if len(args) == 0 {
		return nil, false
	}

	switch method {
	case "DebugContext", "InfoContext", "WarnContext", "ErrorContext":
		if len(args) < 2 {
			return nil, false
		}
		return args[1], true
	case "Log", "LogAttrs":
		if len(args) < 3 {
			return nil, false
		}
		return args[2], true
	default:
		return args[0], true
	}
}

func isSlogCall(pass *analysis.Pass, sel *ast.SelectorExpr) bool {
	// package-level call: slog.Info(...)
	if ident, ok := sel.X.(*ast.Ident); ok {
		if pkgName, ok := pass.TypesInfo.Uses[ident].(*types.PkgName); ok {
			if pkgName.Imported().Path() == "log/slog" {
				return true
			}
		}
	}

	// method call on *slog.Logger or slog.Logger
	selection := pass.TypesInfo.Selections[sel]
	if selection == nil {
		return false
	}
	recv := selection.Recv()
	return typeFromPackage(recv, "log/slog", "Logger")
}

func isZapCall(pass *analysis.Pass, sel *ast.SelectorExpr) bool {
	selection := pass.TypesInfo.Selections[sel]
	if selection == nil {
		return false
	}

	recv := selection.Recv()
	if typeFromPackage(recv, "go.uber.org/zap", "Logger") {
		return true
	}
	if typeFromPackage(recv, "go.uber.org/zap", "SugaredLogger") {
		return true
	}

	return false
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
	if !ok || tv.Value == nil {
		return "", false
	}
	if tv.Value.Kind() != constant.String {
		return "", false
	}
	return constant.StringVal(tv.Value), true
}

func checkLowercase(pass *analysis.Pass, pos token.Pos, msg string) {
	trimmed := strings.TrimSpace(msg)
	if trimmed == "" {
		return
	}

	for _, r := range trimmed {
		if unicode.IsLetter(r) {
			if !unicode.IsLower(r) {
				pass.Reportf(pos, "log message must start with lowercase letter")
			}
			return
		}
	}
}

func checkEnglish(pass *analysis.Pass, pos token.Pos, msg string) {
	for _, r := range msg {
		if !unicode.IsLetter(r) {
			continue
		}
		if r <= unicode.MaxASCII {
			continue
		}
		pass.Reportf(pos, "log message must contain only english text")
		return
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

func checkLiteralSensitive(pass *analysis.Pass, pos token.Pos, msg string) {
	lower := strings.ToLower(msg)
	for _, word := range sensitiveWords {
		if strings.Contains(lower, word+":") || strings.Contains(lower, word+"=") {
			pass.Reportf(pos, "log message may contain sensitive data")
			return
		}
	}
}

func checkSensitiveExpression(pass *analysis.Pass, expr ast.Expr) {
	if !exprContainsDynamicSensitiveConcatenation(pass, expr) {
		return
	}
	pass.Reportf(expr.Pos(), "log message may contain sensitive data")
}

func exprContainsDynamicSensitiveConcatenation(pass *analysis.Pass, expr ast.Expr) bool {
	bin, ok := expr.(*ast.BinaryExpr)
	if !ok || bin.Op != token.ADD {
		return false
	}

	parts := flattenConcat(bin)
	hasSensitiveLiteral := false
	hasNonConstPart := false

	for _, part := range parts {
		if s, ok := constString(pass, part); ok {
			if isSensitivePrefix(strings.ToLower(s)) {
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
	left := flattenConcat(bin.X)
	right := flattenConcat(bin.Y)
	return append(left, right...)
}

func isSensitivePrefix(s string) bool {
	for _, word := range sensitiveWords {
		if strings.Contains(s, word+":") || strings.Contains(s, word+"=") {
			return true
		}
	}
	return false
}

func New(conf any) ([]*analysis.Analyzer, error) {
	if conf != nil {
		switch conf.(type) {
		case map[string]any:
			// reserved for future options
		default:
			return nil, fmt.Errorf("unsupported config type: %T", conf)
		}
	}
	return []*analysis.Analyzer{Analyzer}, nil
}
