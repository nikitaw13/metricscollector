// The staticlint command runs a multichecker over this project's source code.
//
// Analyzers: printf, shadow, structtag (x/tools standard); all SA and ST
// checks from staticcheck.io; qf1001, qf1007 (quickfix); ineffassign;
// bodyclose; and the custom ZeroExitAnalyzer, which forbids os.Exit calls
// in the main function of a main package.
//
// Usage:
//
//	go run ./cmd/staticlint ./...
package main

import (
	"strings"

	"github.com/gordonklaus/ineffassign/pkg/ineffassign"
	"github.com/timakin/bodyclose/passes/bodyclose"
	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/multichecker"
	"golang.org/x/tools/go/analysis/passes/printf"
	"golang.org/x/tools/go/analysis/passes/shadow"
	"golang.org/x/tools/go/analysis/passes/structtag"
	"honnef.co/go/tools/quickfix/qf1001"
	"honnef.co/go/tools/quickfix/qf1007"
	"honnef.co/go/tools/staticcheck"
)

func main() {
	mychecks := []*analysis.Analyzer{
		ZeroExitAnalyzer,
		printf.Analyzer,
		shadow.Analyzer,
		structtag.Analyzer,
		qf1001.Analyzer,
		qf1007.Analyzer,
		ineffassign.Analyzer,
		bodyclose.Analyzer,
	}

	for _, v := range staticcheck.Analyzers {
		if strings.HasPrefix(v.Analyzer.Name, "SA") {
			mychecks = append(mychecks, v.Analyzer)
		}
	}
	for _, v := range staticcheck.Analyzers {
		if strings.HasPrefix(v.Analyzer.Name, "ST") {
			mychecks = append(mychecks, v.Analyzer)
		}
	}

	multichecker.Main(
		mychecks...,
	)
}
