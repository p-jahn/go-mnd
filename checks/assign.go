package checks

import (
	"go/ast"
	"go/token"

	"golang.org/x/tools/go/analysis"

	"github.com/tommy-muehle/go-mnd/v2/config"
)

const AssignCheck = "assign"

type AssignAnalyzer struct {
	pass   *analysis.Pass
	config *config.Config
}

func NewAssignAnalyzer(pass *analysis.Pass, config *config.Config) *AssignAnalyzer {
	return &AssignAnalyzer{
		pass:   pass,
		config: config,
	}
}

func (a *AssignAnalyzer) NodeFilter() []ast.Node {
	return []ast.Node{
		(*ast.KeyValueExpr)(nil),
		(*ast.AssignStmt)(nil),
	}
}

func (a *AssignAnalyzer) Check(n ast.Node) {
	switch expr := n.(type) {
	case *ast.KeyValueExpr:
		switch x := expr.Value.(type) {
		case *ast.BasicLit:
			if a.isMagicNumber(x) {
				a.report(x)
			}
		case *ast.BinaryExpr:
			a.checkBinaryExpr(x)
		}
	case *ast.AssignStmt:
		for _, e := range expr.Rhs {
			switch y := e.(type) {
			case *ast.UnaryExpr:
				a.checkUnaryExpr(y)
			case *ast.BinaryExpr:
				switch x := y.Y.(type) {
				case *ast.UnaryExpr:
					a.checkUnaryExpr(x)
				}
			}
		}
	}
}

func (a *AssignAnalyzer) report(x *ast.BasicLit) {
	a.pass.Reportf(x.Pos(), reportMsg, x.Value, AssignCheck)
}

func (a *AssignAnalyzer) checkUnaryExpr(expr *ast.UnaryExpr) {
	switch x := expr.X.(type) {
	case *ast.BasicLit:
		if a.isMagicNumber(x) {
			a.report(x)
		}
	}
}

func (a *AssignAnalyzer) checkBinaryExpr(expr *ast.BinaryExpr) {
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

func (a *AssignAnalyzer) isMagicNumber(l *ast.BasicLit) bool {
	return (l.Kind == token.FLOAT || l.Kind == token.INT) && !a.config.IsIgnoredNumber(l.Value)
}
