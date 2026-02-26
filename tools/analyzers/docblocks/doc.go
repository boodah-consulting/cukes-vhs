// Package docblocks provides a static analyser that enforces structured documentation
// standards on all exported symbols in Go packages.
//
// # Overview
//
// The docblocks analyser ensures that all exported types, functions, constants,
// and variables have proper godoc comments following the project's documentation
// standards. It integrates with go vet and golangci-lint to provide automated
// documentation quality checks during development and CI.
//
// # Checks Performed
//
// The analyser validates the following documentation requirements:
//
//   - Package-level doc.go files exist with proper package documentation
//   - Exported types have godoc comments starting with the type name
//   - Exported functions have structured documentation sections
//   - Exported constants and variables have descriptive comments
//   - Methods on exported types are properly documented
//
// # Function Documentation Standard
//
// Exported functions should include these structured sections:
//
//	// FunctionName does something important.
//	//
//	// Expected: description of valid input parameters and preconditions.
//	// Returns: description of return values and their meanings.
//	// Side effects: any observable effects beyond the return value.
//	func FunctionName(param Type) (Result, error) { ... }
//
// # Type Documentation Standard
//
// Exported types should have comments starting with the type name:
//
//	// Config holds configuration options for the service.
//	type Config struct { ... }
//
// # Usage
//
// Run via the Makefile:
//
//	make check-docblocks
//
// Or use directly with go vet:
//
//	go vet -vettool=$(which docblocks) ./...
//
// Or integrate with golangci-lint by adding to .golangci.yml:
//
//	linters-settings:
//	  custom:
//	    docblocks:
//	      path: ./tools/analyzers/docblocks/docblocks.so
//	      description: Enforces documentation standards
//
// # CI Integration
//
// The analyser is designed to run as part of the compliance check pipeline:
//
//	make check-compliance  # Includes docblocks check
//
// # Exclusions
//
// The following are excluded from documentation requirements:
//
//   - Unexported (lowercase) symbols
//   - Test files (*_test.go)
//   - Generated code (files with "DO NOT EDIT" header)
//
// # Error Messages
//
// When documentation is missing or malformed, the analyser reports:
//
//	filename.go:10:1: exported type Config should have comment starting with "Config ..."
//	filename.go:20:1: exported function Process should have structured documentation
//
// # Architecture
//
// The analyser is built using the [golang.org/x/tools/go/analysis] framework,
// making it compatible with go vet, golangci-lint, and other analysis tools.
//
// [golang.org/x/tools/go/analysis]: https://pkg.go.dev/golang.org/x/tools/go/analysis
package docblocks
