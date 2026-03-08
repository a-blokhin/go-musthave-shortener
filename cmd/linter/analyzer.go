package main

import (
	"go/ast"
	"go/types"
	"strings"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"
	"golang.org/x/tools/go/ast/inspector"
)

var Analyzer = &analysis.Analyzer{
	Name:     "noprogramexit",
	Doc:      "checks for panic, log.Fatal, and os.Exit usage outside of main function in main package",
	Requires: []*analysis.Analyzer{inspect.Analyzer},
	Run:      run,
}

func run(pass *analysis.Pass) (any, error) {
	// Skip test files and mock files
	for _, f := range pass.Files {
		filename := pass.Fset.File(f.Pos()).Name()
		if strings.HasSuffix(filename, "_test.go") || strings.Contains(filename, "/mocks/") {
			return nil, nil
		}
	}

	inspect := pass.ResultOf[inspect.Analyzer].(*inspector.Inspector)

	nodeFilter := []ast.Node{
		(*ast.CallExpr)(nil),
	}

	inspect.Preorder(nodeFilter, func(n ast.Node) {
		call := n.(*ast.CallExpr)

		// Check for panic
		if isBuiltinCall(pass, call, "panic") {
			pass.Reportf(call.Pos(), "panic should not be used in production code")
			return
		}

		// Check for log.Fatal and os.Exit
		if isFunctionCall(pass, call, "log", "Fatal") || isFunctionCall(pass, call, "log", "Fatalf") || isFunctionCall(pass, call, "log", "Fatalln") {
			if !isInMainFunction(pass, call) {
				pass.Reportf(call.Pos(), "log.Fatal should only be used in main function of main package")
			}
			return
		}

		if isFunctionCall(pass, call, "os", "Exit") {
			if !isInMainFunction(pass, call) {
				pass.Reportf(call.Pos(), "os.Exit should only be used in main function of main package")
			}
			return
		}
	})

	return nil, nil
}

// isBuiltinCall checks if the call is to a builtin function
func isBuiltinCall(pass *analysis.Pass, call *ast.CallExpr, name string) bool {
	fun, ok := call.Fun.(*ast.Ident)
	if !ok {
		return false
	}

	obj := pass.TypesInfo.Uses[fun]
	if obj == nil {
		return false
	}

	_, ok = obj.(*types.Builtin)
	return ok && fun.Name == name
}

// isFunctionCall checks if the call is to a specific package.function
func isFunctionCall(pass *analysis.Pass, call *ast.CallExpr, pkgName, funcName string) bool {
	sel, ok := call.Fun.(*ast.SelectorExpr)
	if !ok {
		return false
	}

	if sel.Sel.Name != funcName {
		return false
	}

	ident, ok := sel.X.(*ast.Ident)
	if !ok {
		return false
	}

	obj := pass.TypesInfo.Uses[ident]
	if obj == nil {
		return false
	}

	pkg, ok := obj.(*types.PkgName)
	return ok && pkg.Imported().Path() == pkgName
}

// isInMainFunction checks if the call is inside the main function of the main package
func isInMainFunction(pass *analysis.Pass, call *ast.CallExpr) bool {
	if pass.Pkg.Name() != "main" {
		return false
	}

	for _, file := range pass.Files {
		if file.Pos() <= call.Pos() && call.Pos() <= file.End() {
			var foundMainFunc *ast.FuncDecl

			ast.Inspect(file, func(n ast.Node) bool {
				if fn, ok := n.(*ast.FuncDecl); ok {
					if fn.Name.Name == "main" && fn.Recv == nil {
						if fn.Pos() <= call.Pos() && call.Pos() <= fn.End() {
							foundMainFunc = fn
							return false
						}
					}
				}
				return true
			})

			return foundMainFunc != nil
		}
	}

	return false
}
