// Package cli implements the cukes-vhs command-line interface.
//
// # Overview
//
// The cli package provides subcommands for managing the complete lifecycle of
// VHS tape generation from Gherkin feature files. It orchestrates the core
// [cukesvhs] library to provide a user-friendly interface for listing scenarios,
// generating tapes, rendering outputs, and managing golden baselines.
//
// # Subcommands
//
// The following subcommands are available:
//
//   - init: Creates a default configuration file for customisation
//   - list: Displays scenarios and their translatability status
//   - generate: Produces VHS tape files from translatable scenarios
//   - run: Executes the full pipeline (generate → render → validate)
//   - update-baseline: Accepts current outputs as new golden baselines
//
// # Usage
//
// The package exposes [Run] as the main entry point:
//
//	os.Exit(cli.Run(os.Args[1:], os.Stdout, os.Stderr))
//
// The Run function returns 0 on success and 1 on error, suitable for
// direct use with os.Exit.
//
// # Version Injection
//
// Use [SetVersion] to configure the version string displayed by --version:
//
//	cli.SetVersion("1.2.3")
//	os.Exit(cli.Run(os.Args[1:], os.Stdout, os.Stderr))
//
// # Common Workflows
//
// List all scenarios with their translation status:
//
//	cukes-vhs list --features features/ --scenarios-dir demos/scenarios/
//
// Generate tapes for all translatable scenarios:
//
//	cukes-vhs generate --all --output /tmp/tapes/
//
// Run the full pipeline with validation:
//
//	cukes-vhs run --all --output /tmp/tapes/ --golden demos/golden/
//
// Update baselines after intentional changes:
//
//	cukes-vhs update-baseline --all --output /tmp/tapes/ --golden demos/golden/
//
// # Feature Directories
//
// The CLI supports two feature sources:
//
//   - --features: Standard Gherkin .feature files from business requirements
//   - --scenarios-dir: VHS-specific scenarios written for demonstrations
//
// Business features generate warnings for untranslatable steps, whilst
// VHS-only scenarios generate errors (they must be fully translatable).
//
// # Configuration
//
// Use 'cukes-vhs init' to create a config/config.tape file that can be
// customised for terminal appearance settings. If no config file exists,
// an embedded default is used automatically.
//
// # Exit Codes
//
//   - 0: Success
//   - 1: Error (invalid arguments, generation failure, validation failure)
//
// [cukesvhs]: github.com/boodah-consulting/cukes-vhs/internal/cukesvhs
package cli
