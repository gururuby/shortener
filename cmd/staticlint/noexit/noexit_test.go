package noexit

import (
	"golang.org/x/tools/go/analysis/analysistest"
	"testing"
)

func TestNoOsExit(t *testing.T) {
	testdata := analysistest.TestData()
	analysistest.Run(t, testdata, NoExitAnalyzer, "./...")
}
