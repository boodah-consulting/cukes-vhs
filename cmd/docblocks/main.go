// Package main provides the entry point for the docblocks analyzer command.
package main

import (
	"golang.org/x/tools/go/analysis/singlechecker"

	"github.com/boodah-consulting/cukes-vhs/tools/analyzers/docblocks"
)

func main() {
	singlechecker.Main(docblocks.Analyzer)
}
