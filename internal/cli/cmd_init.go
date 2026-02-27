package cli

import (
	"fmt"
	"io"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/boodah-consulting/cukes-vhs/internal/cukesvhs"
)

// newInitCmd creates the init command.
func newInitCmd() *cobra.Command {
	var force bool
	var outputDir string

	cmd := &cobra.Command{
		Use:   "init",
		Short: "Initialise config file for customisation",
		Long: `Initialise a cukes-vhs config file for customisation.

Creates a config.tape file that can be customised to change VHS tape styling.
Use --force to overwrite an existing config file.`,
		Example: `  cukes-vhs init                        # Create default config in config/
  cukes-vhs init --output my-config/    # Custom output directory
  cukes-vhs init --force                # Overwrite existing config`,
		RunE: func(cmd *cobra.Command, args []string) error {
			_ = cmd.Context()

			opts := &initOptions{
				force:     force,
				outputDir: outputDir,
			}
			return runInit(opts, cmd.OutOrStdout())
		},
	}

	cmd.Flags().BoolVar(&force, "force", false, "Overwrite existing config file")
	cmd.Flags().StringVar(&outputDir, "output", "config/", "Output directory for config file")

	return cmd
}

// runInit executes the init command.
func runInit(opts *initOptions, out io.Writer) error {
	if err := cliFs.MkdirAll(opts.outputDir, 0o750); err != nil {
		return fmt.Errorf("creating output directory: %w", err)
	}

	configPath := filepath.Join(opts.outputDir, "config.tape")

	if !opts.force {
		if _, err := cliFs.Stat(configPath); err == nil {
			return fmt.Errorf("config file already exists at %s (use --force to overwrite)", configPath)
		}
	}

	content := cukesvhs.DefaultConfig()
	if err := writeFileFs(cliFs, configPath, []byte(content), 0o600); err != nil {
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

// runInitCmd is a legacy wrapper for backward compatibility with tests.
func runInitCmd(args []string, out, errOut io.Writer) int {
	SetWriters(out, errOut)
	cmd := newInitCmd()
	cmd.SetArgs(args)
	cmd.SetOut(out)
	cmd.SetErr(errOut)
	if err := cmd.Execute(); err != nil {
		fmt.Fprintf(errOut, "Error: %v\n", err)
		return 1
	}
	return 0
}

// parseInitFlags parses flags for the init command (legacy interface for tests).
func parseInitFlags(args []string, errOut io.Writer) (*initOptions, error) {
	cmd := newInitCmd()
	cmd.SetArgs(args)
	cmd.SetOut(errOut)
	cmd.SetErr(errOut)

	opts := &initOptions{}

	if err := cmd.ParseFlags(args); err != nil {
		return nil, err
	}

	force, err := cmd.Flags().GetBool("force")
	if err != nil {
		return nil, err
	}
	outputDir, err := cmd.Flags().GetString("output")
	if err != nil {
		return nil, err
	}

	opts.force = force
	opts.outputDir = outputDir

	return opts, nil
}
