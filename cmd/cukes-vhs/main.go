// Package main provides the cukes-vhs CLI entry point.
package main

import (
	"os"

	"github.com/boodah-consulting/cukes-vhs/internal/cli"
)

var version = "dev"

func main() {
	cli.SetVersion(version)
	os.Exit(cli.Execute())
}
