package cli

import (
	"fmt"
	"io"
)

// Run dispatches CLI subcommands based on args.
//
// Expected: args is os.Args[1:]; out and errOut are non-nil writers.
// Returns: 0 on success; 1 on error.
// Side effects: delegates to subcommand handlers which may read/write files.
func Run(args []string, out io.Writer, errOut io.Writer) int {
	if len(args) == 0 {
		printUsageTo(out)
		return 0
	}

	subcommand := args[0]
	rest := args[1:]

	switch subcommand {
	case "init":
		return runInitCmd(rest, out, errOut)
	case "list":
		return runList(rest, out, errOut)
	case "generate":
		return runGenerate(rest, out, errOut)
	case "run":
		return runPipeline(rest, out, errOut)
	case "update-baseline":
		return runUpdateBaseline(rest, out, errOut)
	case "--help", "-h", "help":
		printUsageTo(out)
		return 0
	default:
		fmt.Fprintf(errOut, "Error: unknown subcommand %q\n\n", subcommand)
		printUsageTo(errOut)
		return 1
	}
}

func runInitCmd(args []string, out io.Writer, errOut io.Writer) int {
	opts, err := parseInitFlags(args, errOut)
	if err != nil {
		return 1
	}

	if err := runInit(opts, out); err != nil {
		fmt.Fprintf(errOut, "Error: %v\n", err)
		return 1
	}
	return 0
}

func printUsageTo(out io.Writer) {
	fmt.Fprintln(out, "cukes-vhs — VHS tape generator for Cucumber features")
	fmt.Fprintln(out, "")
	fmt.Fprintln(out, "Usage:")
	fmt.Fprintln(out, "  vhsgen init [flags]              Initialise config file for customisation")
	fmt.Fprintln(out, "  vhsgen list [flags]              List scenarios and their translatability")
	fmt.Fprintln(out, "  vhsgen generate [flags]          Generate VHS tape files from scenarios")
	fmt.Fprintln(out, "  vhsgen run [flags]               Full pipeline: generate → render → validate")
	fmt.Fprintln(out, "  vhsgen update-baseline [flags]   Accept current output as new golden baseline")
	fmt.Fprintln(out, "")
	printInitFlags(out)
	printListFlags(out)
	printGenerateFlags(out)
	printRunFlags(out)
	printUpdateBaselineFlags(out)
	printExamples(out)
}

func printInitFlags(out io.Writer) {
	fmt.Fprintln(out, "init flags:")
	fmt.Fprintln(out, "  --force     Overwrite existing config file")
	fmt.Fprintln(out, "  --output    Output directory (default: config/)")
	fmt.Fprintln(out, "")
}

func printListFlags(out io.Writer) {
	fmt.Fprintln(out, "list flags:")
	fmt.Fprintln(out, "  --features DIR       Directory with .feature files (default: features/)")
	fmt.Fprintln(out, "  --scenarios-dir DIR  Directory with VHS-only .feature files (default: demos/scenarios/)")
	fmt.Fprintln(out, "  --json               Output as JSON")
	fmt.Fprintln(out, "  --count              Show counts broken down by source")
	fmt.Fprintln(out, "  --steps              Show translatable step patterns")
	fmt.Fprintln(out, "")
}

func printGenerateFlags(out io.Writer) {
	fmt.Fprintln(out, "generate flags:")
	fmt.Fprintln(out, "  --all                Generate for all translatable scenarios")
	fmt.Fprintln(out, "  --feature NAME       Filter by feature name")
	fmt.Fprintln(out, "  --scenario NAME      Filter by scenario name")
	fmt.Fprintln(out, "  --features DIR       Directory with .feature files (default: features/)")
	fmt.Fprintln(out, "  --scenarios-dir DIR  Directory with VHS-only .feature files (default: demos/scenarios/)")
	fmt.Fprintln(out, "  --output DIR        Output directory (required)")
	fmt.Fprintln(out, "  --config-source PATH Path to config tape file (default: config/config.tape, falls back to embedded)")
	fmt.Fprintln(out, "  --verbose           Verbose output")
	fmt.Fprintln(out, "  --timeout N         Per-tape render timeout in seconds (default: 120)")
	fmt.Fprintln(out, "  --interactive      Run with interactive TUI")
	fmt.Fprintln(out, "")
}

func printRunFlags(out io.Writer) {
	fmt.Fprintln(out, "run flags:")
	fmt.Fprintln(out, "  --all                Run for all translatable scenarios")
	fmt.Fprintln(out, "  --feature NAME       Filter by feature name")
	fmt.Fprintln(out, "  --scenario NAME      Filter by scenario name")
	fmt.Fprintln(out, "  --features DIR       Directory with .feature files (default: features/)")
	fmt.Fprintln(out, "  --scenarios-dir DIR  Directory with VHS-only .feature files (default: demos/scenarios/)")
	fmt.Fprintln(out, "  --output DIR        Output directory (required)")
	fmt.Fprintln(out, "  --golden DIR        Golden baseline directory (default: demos/golden/)")
	fmt.Fprintln(out, "  --timeout N         Per-tape render timeout in seconds (default: 120)")
	fmt.Fprintln(out, "  --interactive      Run with interactive TUI")
	fmt.Fprintln(out, "  --config-source PATH Path to config tape file (default: config/config.tape, falls back to embedded)")
	fmt.Fprintln(out, "")
}

func printUpdateBaselineFlags(out io.Writer) {
	fmt.Fprintln(out, "update-baseline flags:")
	fmt.Fprintln(out, "  --all                Accept all current outputs as golden baselines")
	fmt.Fprintln(out, "  --golden DIR         Golden baseline directory (default: demos/golden/)")
	fmt.Fprintln(out, "  --output DIR        Output directory containing rendered .ascii files (required)")
	fmt.Fprintln(out, "")
}

func printExamples(out io.Writer) {
	fmt.Fprintln(out, "Examples:")
	fmt.Fprintln(out, "  vhsgen init                              # Create default config")
	fmt.Fprintln(out, "  vhsgen init --output my-config/         # Custom output directory")
	fmt.Fprintln(out, "  vhsgen list --features features/ --scenarios-dir demos/scenarios/")
	fmt.Fprintln(out, "  vhsgen list --json")
	fmt.Fprintln(out, "  vhsgen list --count")
	fmt.Fprintln(out, "  vhsgen list --steps")
	fmt.Fprintln(out, "  vhsgen list --steps --json")
	fmt.Fprintln(out, "  vhsgen generate --all --features features/ --scenarios-dir demos/scenarios/ --output /tmp/tapes/")
	fmt.Fprintln(out, "  vhsgen generate --feature onboarding --output /tmp/test/")
	fmt.Fprintln(out, "  vhsgen generate --interactive --all --output /tmp/tapes/   # Interactive mode")
	fmt.Fprintln(out, "  vhsgen run --all --output /tmp/tapes/ --golden demos/golden/")
	fmt.Fprintln(out, "  vhsgen run --feature onboarding --output /tmp/tapes/")
	fmt.Fprintln(out, "  vhsgen update-baseline --all --output /tmp/tapes/ --golden demos/golden/")
	fmt.Fprintln(out, "  vhsgen update-baseline my-scenario --output /tmp/tapes/")
}
