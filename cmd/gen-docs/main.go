// Package main provides a tool for generating cukes-vhs documentation.
//
// This tool generates man pages from the Cobra command tree using
// the cobra/doc package. Run it via 'make man' to regenerate the
// man pages in docs/man/.
package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra/doc"

	"github.com/boodah-consulting/cukes-vhs/internal/cli"
)

const (
	manSection = "1"
	manDir     = "docs/man"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
	outDir := manDir
	if len(os.Args) > 1 {
		outDir = os.Args[1]
	}

	if err := os.MkdirAll(outDir, 0o750); err != nil {
		return fmt.Errorf("creating output directory: %w", err)
	}

	cmd := cli.NewRootCmd()

	header := &doc.GenManHeader{
		Title:   "CUKES-VHS",
		Section: manSection,
		Source:  "cukes-vhs",
		Manual:  "cukes-vhs Manual",
	}

	if err := doc.GenManTree(cmd, header, outDir); err != nil {
		return fmt.Errorf("generating man pages: %w", err)
	}

	absPath, _ := filepath.Abs(outDir)
	fmt.Printf("Man pages generated in %s\n", absPath)

	return nil
}
