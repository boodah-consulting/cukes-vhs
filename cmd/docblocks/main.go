// Package main provides the entry point for the docblocks analyzer command.
package main

import (
	"github.com/boodah-consulting/cukesvhs/tools/analyzers/docblocks"
	"golang.org/x/tools/go/analysis/singlechecker"
)

func main() {
	singlechecker.Main(docblocks.Analyzer)
}
