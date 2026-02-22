// Package main provides the cukes-vhs CLI entry point.
package main

import (
	"os"

	"github.com/boodah-consulting/cukes-vhs/internal/cli"
)

func main() {
	os.Exit(cli.Run(os.Args[1:], os.Stdout, os.Stderr))
}
