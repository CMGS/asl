// Package cmpor flags zero-value fallbacks written as an if statement that cmp.Or expresses in one call; -fix rewrites them (run goimports after for the cmp import).
package cmpor

import (
	"go/ast"
	"go/constant"
	"go/token"
	"go/types"

	"golang.org/x/tools/go/analysis"

	"github.com/CMGS/asl/internal/source"
)

var (
	Analyzer = newAnalyzer()

	// signed widens the x > 0 form to signed numbers, where cmp.Or changes what a negative x means; advisory runs only.
	signed bool
)

func newAnalyzer() *analysis.Analyzer {
	a := &analysis.Analyzer{
		Name: "cmpor",
		Doc:  "flag zero-value fallbacks (if x != zero { return x }; return y) that cmp.Or expresses",
		Run:  run,
	}
	a.Flags.BoolVar(&signed, "signed", false, "also report x > 0 fallbacks on signed numbers")
	return a
}

func run(pass *analysis.Pass) (any, error) {
	for f := range source.Files(pass) {
		for n := range ast.Preorder(f) {
			if block, ok := n.(*ast.BlockStmt); ok {
				checkBlock(pass, block.List)
			}
		}
	}
	return nil, nil
}

func checkBlock(pass *analysis.Pass, stmts []ast.Stmt) {
	src := func(e ast.Expr) string {
		f := pass.Fset.File(e.Pos())
		data, err := pass.ReadFile(f.Name())
		if err != nil {
			return types.ExprString(e)
		}
		return string(data[f.Offset(e.Pos()):f.Offset(e.End())])
	}
	report := func(pos, end token.Pos, msg, fix string) {
		pass.Report(analysis.Diagnostic{
			Pos:     pos,
			Message: msg,
			SuggestedFixes: []analysis.SuggestedFix{{
				Message:   "use cmp.Or",
				TextEdits: []analysis.TextEdit{{Pos: pos, End: end, NewText: []byte(fix)}},
			}},
		})
	}
	for i, stmt := range stmts {
		ifs, ok := stmt.(*ast.IfStmt)
		if !ok || ifs.Init != nil || ifs.Else != nil || len(ifs.Body.List) != 1 {
			continue
		}
		x, nonZero, ok := fallbackCond(pass, ifs.Cond)
		if !ok {
			continue
		}
		name := types.ExprString(x)
		switch body := ifs.Body.List[0].(type) {
		case *ast.ReturnStmt:
			if i+1 >= len(stmts) {
				continue
			}
			next, ok := stmts[i+1].(*ast.ReturnStmt)
			if !ok || len(body.Results) != 1 || len(next.Results) != 1 {
				continue
			}
			kept, fallback := body.Results[0], next.Results[0]
			if !nonZero {
				kept, fallback = fallback, kept
			}
			if types.ExprString(kept) == name && pure(fallback) {
				call := "cmp.Or(" + src(x) + ", " + src(fallback) + ")"
				report(ifs.Pos(), next.End(), "if/return fallback on "+name+" is "+call, "return "+call)
			}
		case *ast.AssignStmt:
			if nonZero || body.Tok != token.ASSIGN || len(body.Lhs) != 1 || len(body.Rhs) != 1 || types.ExprString(body.Lhs[0]) != name || !pure(body.Rhs[0]) {
				continue
			}
			call := "cmp.Or(" + src(x) + ", " + src(body.Rhs[0]) + ")"
			report(ifs.Pos(), ifs.End(), "zero-value fallback on "+name+" is "+name+" = "+call, src(x)+" = "+call)
		}
	}
}

// fallbackCond recognizes x != zero, x == zero and x > 0 (unsigned, or signed with -signed) for a comparable x.
func fallbackCond(pass *analysis.Pass, cond ast.Expr) (x ast.Expr, nonZero, ok bool) {
	bin, isBin := cond.(*ast.BinaryExpr)
	if !isBin {
		return nil, false, false
	}
	x, zero := bin.X, bin.Y
	if isZero(pass, x) {
		x, zero = bin.Y, bin.X
	}
	if !isZero(pass, zero) || !isOperand(x) {
		return nil, false, false
	}
	t := pass.TypesInfo.TypeOf(x)
	if t == nil || !types.Comparable(t) {
		return nil, false, false
	}
	switch bin.Op {
	case token.NEQ:
		return x, true, true
	case token.EQL:
		return x, false, true
	case token.GTR:
		return x, true, x == bin.X && numericFallback(t)
	}
	return nil, false, false
}

func numericFallback(t types.Type) bool {
	b, ok := t.Underlying().(*types.Basic)
	if !ok {
		return false
	}
	return b.Info()&types.IsUnsigned != 0 || (signed && b.Info()&types.IsNumeric != 0)
}

// pure reports whether e has no call or func literal: cmp.Or evaluates every argument, so a fallback with effects must stay lazy.
func pure(e ast.Expr) bool {
	for n := range ast.Preorder(e) {
		switch n.(type) {
		case *ast.CallExpr, *ast.FuncLit:
			return false
		}
	}
	return true
}

func isOperand(e ast.Expr) bool {
	switch e := e.(type) {
	case *ast.Ident:
		return e.Name != "nil"
	case *ast.SelectorExpr:
		return isOperand(e.X)
	case *ast.ParenExpr:
		return isOperand(e.X)
	}
	return false
}

func isZero(pass *analysis.Pass, e ast.Expr) bool {
	if id, ok := e.(*ast.Ident); ok && id.Name == "nil" {
		_, isNil := pass.TypesInfo.Uses[id].(*types.Nil)
		return isNil
	}
	tv, ok := pass.TypesInfo.Types[e]
	if !ok || tv.Value == nil {
		return false
	}
	switch tv.Value.Kind() {
	case constant.String:
		return constant.StringVal(tv.Value) == ""
	case constant.Int, constant.Float:
		return constant.Sign(tv.Value) == 0
	}
	return false
}
