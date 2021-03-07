package checks

import (
	"go/ast"
	"go/token"

	"golang.org/x/tools/go/analysis"

	"github.com/tommy-muehle/go-mnd/v2/config"
)

const ReturnCheck = "return"

type ReturnAnalyzer struct {
	pass   *analysis.Pass
	config *config.Config
}

func NewReturnAnalyzer(pass *analysis.Pass, config *config.Config) *ReturnAnalyzer {
	return &ReturnAnalyzer{
		pass:   pass,
		config: config,
	}
}

func (a *ReturnAnalyzer) NodeFilter() []ast.Node {
	return []ast.Node{
		(*ast.ReturnStmt)(nil),
	}
}

func (a *ReturnAnalyzer) Check(n ast.Node) {
	stmt, ok := n.(*ast.ReturnStmt)
	if !ok {
		return
	}

	for _, expr := range stmt.Results {
		switch x := expr.(type) {
		case *ast.BasicLit:
			if a.isMagicNumber(x) {
				a.report(x)
			}
		case *ast.BinaryExpr:
			a.checkBinaryExpr(x)
		}
	}
}

func (a *ReturnAnalyzer) report(x *ast.BasicLit) {
	a.pass.Reportf(x.Pos(), reportMsg, x.Value, CaseCheck)
}

func (a *ReturnAnalyzer) checkBinaryExpr(expr *ast.BinaryExpr) {
	switch x := expr.X.(type) {
	case *ast.BasicLit:
		if a.isMagicNumber(x) {
			a.report(x)
		}
	}

	switch y := expr.Y.(type) {
	case *ast.BasicLit:
		if a.isMagicNumber(y) {
			a.report(y)
		}
	}
}

func (a *ReturnAnalyzer) isMagicNumber(l *ast.BasicLit) bool {
	return (l.Kind == token.FLOAT || l.Kind == token.INT) && !a.config.IsIgnoredNumber(l.Value)
}
