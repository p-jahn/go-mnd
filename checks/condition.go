package checks

import (
	"go/ast"
	"go/token"

	"golang.org/x/tools/go/analysis"

	"github.com/tommy-muehle/go-mnd/v2/config"
)

const ConditionCheck = "condition"

type ConditionAnalyzer struct {
	pass   *analysis.Pass
	config *config.Config
}

func NewConditionAnalyzer(pass *analysis.Pass, config *config.Config) *ConditionAnalyzer {
	return &ConditionAnalyzer{
		pass:   pass,
		config: config,
	}
}

func (a *ConditionAnalyzer) NodeFilter() []ast.Node {
	return []ast.Node{
		(*ast.IfStmt)(nil),
	}
}

func (a *ConditionAnalyzer) Check(n ast.Node) {
	expr, ok := n.(*ast.IfStmt).Cond.(*ast.BinaryExpr)
	if !ok {
		return
	}

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

func (a *ConditionAnalyzer) report(x *ast.BasicLit) {
	a.pass.Reportf(x.Pos(), reportMsg, x.Value, CaseCheck)
}

func (a *ConditionAnalyzer) isMagicNumber(l *ast.BasicLit) bool {
	return (l.Kind == token.FLOAT || l.Kind == token.INT) && !a.config.IsIgnoredNumber(l.Value)
}
