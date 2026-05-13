package main

import (
	"golang.org/x/tools/go/analysis/singlechecker"

	"github.com/Arturikou/urlshortener/cmd/linter/analyzer"
)

func main() {
	singlechecker.Main(analyzer.Analyzer)
}
