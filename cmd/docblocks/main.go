// Package main provides the entry point for the docblocks analyzer command.
package main

import (
	"github.com/baphled/cukes-vhs/tools/analyzers/docblocks"
	"golang.org/x/tools/go/analysis/singlechecker"
)

func main() {
	singlechecker.Main(docblocks.Analyzer)
}
