// Package noexit provides a static analysis tool that forbids direct calls to os.Exit
// in the main function of the main package.
//
// The analyzer helps enforce better program termination practices by requiring
// proper error handling and cleanup before program exit.
//
// Usage:
//
// To use this analyzer with go vet:
//
//	go vet -vettool=$(which noexit) ./...
//
// Or as part of your analysis suite:
//
//	analyzer := noexit.NoExitAnalyzer
//
// Example violation:
//
//	package main
//
//	import "os"
//
//	func main() {
//	    os.Exit(1) // will be flagged by the analyzer
//	}
//
// The analyzer will report:
//
//	main.go:6:2: direct call to os.Exit in main function of main package is forbidden
package noexit

import (
	"go/ast"
	"strings"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"
)

// Analyzer is the analyzer variable that checks for forbidden os.Exit calls.
// It implements the analysis.Analyzer interface and can be used with analysis tools.
//
// The analyzer checks all files in the main package named "main.go" for a main()
// function containing direct calls to os.Exit().
var Analyzer = &analysis.Analyzer{
	Name:     "noexit",
	Doc:      "forbid direct calls to os.Exit in main function of main package",
	Requires: []*analysis.Analyzer{inspect.Analyzer},
	Run:      run,
}

// run is the analysis function that implements the check logic.
// It examines each file in the package, looking for main packages and checking
// their main functions for os.Exit calls.
func run(pass *analysis.Pass) (interface{}, error) {
	for _, file := range pass.Files {

		// Ignore cache go-build files
		if strings.Contains(pass.Fset.File(file.Pos()).Name(), "/go-build/") {
			continue
		}

		if pass.Pkg.Name() == "main" {
			ast.Inspect(file, func(n ast.Node) bool {
				// Find main function
				if fn, ok := n.(*ast.FuncDecl); ok && fn.Name.Name == "main" {
					// Detect os.Exit
					ast.Inspect(fn.Body, func(n ast.Node) bool {
						call, ok := n.(*ast.CallExpr)
						if !ok {
							return true
						}

						sel, ok := call.Fun.(*ast.SelectorExpr)
						if !ok {
							return true
						}

						ident, ok := sel.X.(*ast.Ident)

						if ok && ident.Name == "os" && sel.Sel.Name == "Exit" {
							pass.Reportf(call.Pos(), "direct call to os.Exit in main function of main package is forbidden")
						}

						return true
					})
				}
				return true
			})
		}
	}
	return nil, nil
}
