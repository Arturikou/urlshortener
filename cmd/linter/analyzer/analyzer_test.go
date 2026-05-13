package analyzer_test

import (
	"testing"

	"golang.org/x/tools/go/analysis/analysistest"

	"github.com/Arturikou/urlshortener/cmd/linter/analyzer"
)

func TestPanicCheck(t *testing.T) {
	analysistest.Run(t, analysistest.TestData(), analyzer.Analyzer, "nopanic")
}

func TestExitCheck(t *testing.T) {
	analysistest.Run(t, analysistest.TestData(), analyzer.Analyzer, "noexit")
}
