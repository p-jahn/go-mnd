package checks

import (
	"go/ast"
	"go/token"
	"strconv"
	"sync"

	"golang.org/x/tools/go/analysis"

	"github.com/tommy-muehle/go-mnd/v2/config"
)

const ArgumentCheck = "argument"

type ArgumentAnalyzer struct {
	config *config.Config
	pass   *analysis.Pass

	mu             sync.RWMutex
	constPositions map[string]bool
}

func NewArgumentAnalyzer(pass *analysis.Pass, config *config.Config) *ArgumentAnalyzer {
	return &ArgumentAnalyzer{
		pass:   pass,
		config: config,

		constPositions: make(map[string]bool),
	}
}

func (a *ArgumentAnalyzer) NodeFilter() []ast.Node {
	return []ast.Node{
		(*ast.GenDecl)(nil),
		(*ast.CallExpr)(nil),
	}
}

func (a *ArgumentAnalyzer) Check(n ast.Node) {
	switch expr := n.(type) {
	case *ast.CallExpr:
		a.checkCallExpr(expr)
	case *ast.GenDecl:
		if expr.Tok == token.CONST {
			pos := a.pass.Fset.Position(expr.TokPos)

			a.mu.Lock()
			a.constPositions[constPosKey(pos)] = true
			a.mu.Unlock()
		}
	}
}

func (a *ArgumentAnalyzer) report(x *ast.BasicLit) {
	a.pass.Reportf(x.Pos(), reportMsg, x.Value, ArgumentCheck)
}

func (a *ArgumentAnalyzer) checkCallExpr(expr *ast.CallExpr) {
	pos := a.pass.Fset.Position(expr.Pos())

	a.mu.RLock()
	isDefinedAsConstant := a.constPositions[constPosKey(pos)]
	a.mu.RUnlock()

	if isDefinedAsConstant {
		return
	}

	if f, isSelectorExpr := expr.Fun.(*ast.SelectorExpr); isSelectorExpr {
		prefix, isIdentifier := f.X.(*ast.Ident)
		if isIdentifier && a.config.IsIgnoredFunction(prefix.Name+"."+f.Sel.Name) {
			return
		}
	}

	for i, arg := range expr.Args {
		switch x := arg.(type) {
		case *ast.BasicLit:
			if !a.isMagicNumber(x) {
				continue
			}
			// If it's a magic number and has no previous element, report it
			if i == 0 {
				a.report(x)
				continue
			}

			// Otherwise check the previous element type
			_, isChannel := expr.Args[i-1].(*ast.ChanType)
			if isChannel && a.isMagicNumber(x) {
				a.report(x)
			}
		case *ast.BinaryExpr:
			a.checkBinaryExpr(x)
		}
	}
}

func (a *ArgumentAnalyzer) checkBinaryExpr(expr *ast.BinaryExpr) {
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

func (a *ArgumentAnalyzer) isMagicNumber(l *ast.BasicLit) bool {
	return (l.Kind == token.FLOAT || l.Kind == token.INT) && !a.config.IsIgnoredNumber(l.Value)
}

func constPosKey(pos token.Position) string {
	return pos.Filename + ":" + strconv.Itoa(pos.Line)
}
