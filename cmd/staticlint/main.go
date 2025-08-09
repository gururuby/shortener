// Package main implements a custom static analysis tool that combines multiple Go analyzers
// into a single executable. It includes standard go/analysis passes, selected staticcheck
// analyzers, style checks, and custom analyzers like the noexit checker.
//
// The tool is designed to enforce code quality standards and catch potential issues by running
// multiple analyzers simultaneously through the multichecker framework.
//
// # Included Analyzers
//
// The tool combines several categories of analyzers:
//
// 1. Standard go/analysis passes:
//    - asmdecl, assign, atomic, bools, buildtag, cgocall, composite, copylock
//    - errorsas, fieldalignment, httpresponse, loopclosure, lostcancel
//    - nilfunc, printf, shadow, shift, sigchanyzer, sortslice, stdmethods
//    - stringintconv, structtag, testinggoroutine, tests, unmarshal
//    - unreachable, unsafeptr, unusedresult
//
// 2. Staticcheck analyzers (SA* series):
//    - All analyzers from staticcheck with names starting with "SA"
//
// 3. Stylecheck analyzers:
//    - ST1000: Checks for incorrect or missing package comments
//    - ST1001: Enforces naming style conventions
//
// 4. Custom analyzers:
//    - noexit: Forbids direct calls to os.Exit in main functions
//
// # Usage
//
// To run the analyzer:
//   go build -o staticlint && ./staticlint ./...
//
// Alternatively, use with go vet:
//   go vet -vettool=staticlint ./...
//
// # Configuration
//
// The analyzers can be configured by:
// - Adding or removing analyzers from the checks slice in main()
// - Using analyzer-specific configuration files where supported
//
// # Output
//
// The tool outputs findings from all enabled analyzers, with each finding
// including:
// - File position
// - Analyzer name
// - Diagnostic message
//
// # Extending
//
// To add new analyzers:
// 1. Import the analyzer package
// 2. Append its Analyzer to the checks slice in main()

package main

import (
	"github.com/gururuby/shortener/cmd/staticlint/noexit"
	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/multichecker"
	"golang.org/x/tools/go/analysis/passes/asmdecl"
	"golang.org/x/tools/go/analysis/passes/assign"
	"golang.org/x/tools/go/analysis/passes/atomic"
	"golang.org/x/tools/go/analysis/passes/bools"
	"golang.org/x/tools/go/analysis/passes/buildtag"
	"golang.org/x/tools/go/analysis/passes/cgocall"
	"golang.org/x/tools/go/analysis/passes/composite"
	"golang.org/x/tools/go/analysis/passes/copylock"
	"golang.org/x/tools/go/analysis/passes/errorsas"
	"golang.org/x/tools/go/analysis/passes/fieldalignment"
	"golang.org/x/tools/go/analysis/passes/httpresponse"
	"golang.org/x/tools/go/analysis/passes/loopclosure"
	"golang.org/x/tools/go/analysis/passes/lostcancel"
	"golang.org/x/tools/go/analysis/passes/nilfunc"
	"golang.org/x/tools/go/analysis/passes/printf"
	"golang.org/x/tools/go/analysis/passes/shadow"
	"golang.org/x/tools/go/analysis/passes/shift"
	"golang.org/x/tools/go/analysis/passes/sigchanyzer"
	"golang.org/x/tools/go/analysis/passes/sortslice"
	"golang.org/x/tools/go/analysis/passes/stdmethods"
	"golang.org/x/tools/go/analysis/passes/stringintconv"
	"golang.org/x/tools/go/analysis/passes/structtag"
	"golang.org/x/tools/go/analysis/passes/testinggoroutine"
	"golang.org/x/tools/go/analysis/passes/tests"
	"golang.org/x/tools/go/analysis/passes/unmarshal"
	"golang.org/x/tools/go/analysis/passes/unreachable"
	"golang.org/x/tools/go/analysis/passes/unsafeptr"
	"golang.org/x/tools/go/analysis/passes/unusedresult"
	"honnef.co/go/tools/staticcheck"
	"honnef.co/go/tools/stylecheck/st1000"
	"honnef.co/go/tools/stylecheck/st1001"
)

func main() {
	var checks []*analysis.Analyzer

	checks = append(checks,
		asmdecl.Analyzer,
		assign.Analyzer,
		atomic.Analyzer,
		bools.Analyzer,
		buildtag.Analyzer,
		cgocall.Analyzer,
		composite.Analyzer,
		copylock.Analyzer,
		errorsas.Analyzer,
		fieldalignment.Analyzer,
		httpresponse.Analyzer,
		loopclosure.Analyzer,
		lostcancel.Analyzer,
		nilfunc.Analyzer,
		printf.Analyzer,
		shadow.Analyzer,
		shift.Analyzer,
		sigchanyzer.Analyzer,
		sortslice.Analyzer,
		stdmethods.Analyzer,
		stringintconv.Analyzer,
		structtag.Analyzer,
		testinggoroutine.Analyzer,
		tests.Analyzer,
		unmarshal.Analyzer,
		unreachable.Analyzer,
		unsafeptr.Analyzer,
		unusedresult.Analyzer,
	)

	for _, v := range staticcheck.Analyzers {
		if v.Analyzer.Name[0:2] == "SA" {
			checks = append(checks, v.Analyzer)
		}
	}

	checks = append(checks,
		st1000.SCAnalyzer.Analyzer, // Incorrect or missing package comment
		st1001.SCAnalyzer.Analyzer, // Naming style
	)

	checks = append(checks, noexit.Analyzer)

	multichecker.Main(checks...)
}
