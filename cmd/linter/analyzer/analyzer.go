package analyzer

import (
	"go/ast"
	"go/token"
	"strings"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"
	"golang.org/x/tools/go/ast/inspector"
)

var Analyzer = &analysis.Analyzer{
	Name:     "linter",
	Doc:      "reports use of panic and calls to os.Exit/log.Fatal outside of main() in package main",
	Requires: []*analysis.Analyzer{inspect.Analyzer},
	Run:      run,
}

func run(pass *analysis.Pass) (interface{}, error) {
	insp := pass.ResultOf[inspect.Analyzer].(*inspector.Inspector)

	insp.Preorder([]ast.Node{(*ast.FuncDecl)(nil)}, func(n ast.Node) {
		fn := n.(*ast.FuncDecl)
		if isGenerated(fileFor(pass, fn.Pos())) {
			return
		}
		inMainFunc := pass.Pkg.Name() == "main" && fn.Name.Name == "main"

		ast.Inspect(fn.Body, func(node ast.Node) bool {
			call, ok := node.(*ast.CallExpr)
			if !ok {
				return true
			}
			checkPanic(pass, call)
			if !inMainFunc {
				checkExit(pass, call)
			}
			return true
		})
	})
	return nil, nil
}

func checkPanic(pass *analysis.Pass, call *ast.CallExpr) {
	ident, ok := call.Fun.(*ast.Ident)
	if !ok || ident.Name != "panic" {
		return
	}
	pass.Reportf(call.Pos(), "use of built-in panic")
}

func checkExit(pass *analysis.Pass, call *ast.CallExpr) {
	sel, ok := call.Fun.(*ast.SelectorExpr)
	if !ok {
		return
	}
	pkg, ok := sel.X.(*ast.Ident)
	if !ok {
		return
	}
	if (pkg.Name == "os" && sel.Sel.Name == "Exit") ||
		(pkg.Name == "log" && strings.HasPrefix(sel.Sel.Name, "Fatal")) {
		pass.Reportf(call.Pos(), "call to %s.%s outside main() of package main", pkg.Name, sel.Sel.Name)
	}
}

func isGenerated(file *ast.File) bool {
	if file == nil {
		return false
	}
	for _, cg := range file.Comments {
		for _, c := range cg.List {
			if strings.HasPrefix(c.Text, "// Code generated") {
				return true
			}
		}
	}
	return false
}

func fileFor(pass *analysis.Pass, pos token.Pos) *ast.File {
	for _, f := range pass.Files {
		if f.FileStart <= pos && pos <= f.FileEnd {
			return f
		}
	}
	return nil
}
