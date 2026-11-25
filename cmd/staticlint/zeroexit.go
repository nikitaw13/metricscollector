package main

import (
	"go/ast"
	"go/types"
	"strings"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"
	"golang.org/x/tools/go/ast/inspector"
)

// ZeroExitAnalyzer forbids direct os.Exit calls from the main function of
// a main package; early exits bypass deferred cleanup.
var ZeroExitAnalyzer = &analysis.Analyzer{
	Name:     "zeroexitcheck",
	Doc:      "check for using os.Exit in main",
	Run:      run,
	Requires: []*analysis.Analyzer{inspect.Analyzer},
}

func run(pass *analysis.Pass) (any, error) {
	if pass.Pkg.Name() != "main" {
		return nil, nil
	}
	if strings.HasSuffix(pass.Pkg.Path(), ".test") {
		return nil, nil
	}

	inspect := pass.ResultOf[inspect.Analyzer].(*inspector.Inspector)

	nodeFilter := []ast.Node{
		(*ast.FuncDecl)(nil),
	}

	inspect.Preorder(nodeFilter, func(n ast.Node) {
		fd, ok := n.(*ast.FuncDecl)
		if !ok || fd.Recv != nil || fd.Name.Name != "main" {
			return
		}

		ast.Inspect(fd.Body, func(n ast.Node) bool {
			call, ok := n.(*ast.CallExpr)
			if !ok {
				return true
			}
			sel, ok := call.Fun.(*ast.SelectorExpr)
			if !ok {
				return true
			}
			pkgIdent, ok := sel.X.(*ast.Ident)
			if !ok {
				return true
			}
			pkgName, ok := pass.TypesInfo.Uses[pkgIdent].(*types.PkgName)
			if ok && pkgName.Imported().Path() == "os" && sel.Sel.Name == "Exit" {
				pass.Reportf(call.Pos(), "os.Exit call in main function is forbidden")
			}
			return true
		})
	})
	return nil, nil
}
