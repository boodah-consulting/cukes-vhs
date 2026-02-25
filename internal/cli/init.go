package cli

import (
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/boodah-consulting/cukesvhs/internal/cukesvhs"
)

// runInit executes the init command.
func runInit(opts *initOptions, out io.Writer) error {
	// Ensure output directory exists
	if err := os.MkdirAll(opts.outputDir, 0o750); err != nil {
		return fmt.Errorf("creating output directory: %w", err)
	}

	configPath := filepath.Join(opts.outputDir, "config.tape")

	// Check if file exists
	if !opts.force {
		if _, err := os.Stat(configPath); err == nil {
			return fmt.Errorf("config file already exists at %s (use --force to overwrite)", configPath)
		}
	}

	// Get embedded config and write to file
	content := cukesvhs.DefaultConfig()
	if err := os.WriteFile(configPath, []byte(content), 0o600); err != nil {
		return fmt.Errorf("writing config file: %w", err)
	}

	fmt.Fprintf(out, "Created config file: %s\n", configPath)
	fmt.Fprintf(out, "You can customise this file to change the VHS tape styling.\n")
	fmt.Fprintf(out, "Then use: cukes-vhs generate --config-source %s ...\n", configPath)

	return nil
}

// initOptions holds flags for the init command.
type initOptions struct {
	force     bool
	outputDir string
}

// parseInitFlags parses flags for the init command.
func parseInitFlags(args []string, errOut io.Writer) (*initOptions, error) {
	fs := flag.NewFlagSet("init", flag.ContinueOnError)
	fs.SetOutput(errOut)

	opts := &initOptions{}
	fs.BoolVar(&opts.force, "force", false, "Overwrite existing config file")
	fs.StringVar(&opts.outputDir, "output", "config/", "Output directory for config file")

	if err := fs.Parse(args); err != nil {
		return nil, err
	}

	return opts, nil
}
