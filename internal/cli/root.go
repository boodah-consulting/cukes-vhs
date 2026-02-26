package cli

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/spf13/afero"
	"github.com/spf13/cobra"

	"github.com/boodah-consulting/cukes-vhs/internal/cukesvhs"
)

var (
	cliVersion = "dev"
	rootCmd    *cobra.Command
	outWriter  io.Writer = os.Stdout
	errWriter  io.Writer = os.Stderr
	cliFs      afero.Fs  = afero.NewOsFs()
)

// SetVersion sets the version string to be displayed by --version.
func SetVersion(v string) {
	cliVersion = v
}

// SetWriters allows tests to override stdout and stderr.
func SetWriters(out, errOut io.Writer) {
	outWriter = out
	errWriter = errOut
}

// SetFs allows tests to override the filesystem.
func SetFs(fs afero.Fs) {
	cliFs = fs
	cukesvhs.SetDefaultFs(fs)
}

// NewRootCmd creates the root command for cukes-vhs.
// It is exported to allow documentation generation tools to access the command tree.
func NewRootCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "cukes-vhs",
		Short: "VHS tape generator for Cucumber features",
		Long: `cukes-vhs — VHS tape generator for Cucumber features

cukes-vhs converts Gherkin BDD scenarios into VHS tape files for
automated terminal recordings using charmbracelet/vhs.`,
		Version:       cliVersion,
		SilenceUsage:  true,
		SilenceErrors: true,
	}

	cmd.SetOut(outWriter)
	cmd.SetErr(errWriter)

	cmd.AddCommand(newInitCmd())
	cmd.AddCommand(newListCmd())
	cmd.AddCommand(newGenerateCmd())
	cmd.AddCommand(newRunCmd())
	cmd.AddCommand(newUpdateBaselineCmd())
	cmd.AddCommand(newCompletionCmd())

	return cmd
}

// Execute runs the root command with signal handling for graceful shutdown.
func Execute() int {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	rootCmd = NewRootCmd()
	if err := rootCmd.ExecuteContext(ctx); err != nil {
		return 1
	}
	return 0
}

// Run dispatches CLI subcommands based on args (legacy interface for tests).
func Run(args []string, out io.Writer, errOut io.Writer) int {
	SetWriters(out, errOut)
	rootCmd = NewRootCmd()
	rootCmd.SetArgs(args)
	rootCmd.SetOut(out)
	rootCmd.SetErr(errOut)

	if len(args) > 0 {
		subcommand := args[0]
		found := false
		for _, cmd := range rootCmd.Commands() {
			if cmd.Name() == subcommand {
				found = true
				break
			}
		}
		if !found && !isFlag(subcommand) {
			fmt.Fprintf(errOut, "Error: unknown subcommand %q\n\n", subcommand)
			return 1
		}
	}

	if err := rootCmd.Execute(); err != nil {
		// Format error message without extra "Error:" prefix for errors that already
		// contain structured error text
		errMsg := err.Error()
		if !startsWithErrorPrefix(errMsg) {
			fmt.Fprintf(errOut, "Error: %v\n", err)
		}
		return 1
	}
	return 0
}

// isFlag checks if the argument is a flag.
func isFlag(arg string) bool {
	return len(arg) > 0 && arg[0] == '-'
}

// startsWithErrorPrefix checks if the error message already starts with "Error".
func startsWithErrorPrefix(s string) bool {
	return strings.HasPrefix(s, "Error")
}

// newCompletionCmd creates the completion command for shell completion scripts.
func newCompletionCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "completion [bash|zsh|fish|powershell]",
		Short: "Generate shell completion scripts",
		Long: `Generate shell completion scripts for cukes-vhs.

To load completions:

Bash:
  $ source <(cukes-vhs completion bash)
  # To load completions for each session, execute once:
  # Linux:
  $ cukes-vhs completion bash > /etc/bash_completion.d/cukes-vhs
  # macOS:
  $ cukes-vhs completion bash > $(brew --prefix)/etc/bash_completion.d/cukes-vhs

Zsh:
  # If shell completion is not already enabled in your environment,
  # you will need to enable it. You can execute the following once:
  $ echo "autoload -U compinit; compinit" >> ~/.zshrc
  # To load completions for each session, execute once:
  $ cukes-vhs completion zsh > "${fpath[1]}/_cukes-vhs"
  # You will need to start a new shell for this setup to take effect.

Fish:
  $ cukes-vhs completion fish | source
  # To load completions for each session, execute once:
  $ cukes-vhs completion fish > ~/.config/fish/completions/cukes-vhs.fish

PowerShell:
  PS> cukes-vhs completion powershell | Out-String | Invoke-Expression
  # To load completions for every new session, run:
  PS> cukes-vhs completion powershell > cukes-vhs.ps1
  # and source this file from your PowerShell profile.
`,
		DisableFlagsInUseLine: true,
		ValidArgs:             []string{"bash", "zsh", "fish", "powershell"},
		Args:                  cobra.MatchAll(cobra.ExactArgs(1), cobra.OnlyValidArgs),
		RunE: func(cmd *cobra.Command, args []string) error {
			switch args[0] {
			case "bash":
				return cmd.Root().GenBashCompletion(cmd.OutOrStdout())
			case "zsh":
				return cmd.Root().GenZshCompletion(cmd.OutOrStdout())
			case "fish":
				return cmd.Root().GenFishCompletion(cmd.OutOrStdout(), true)
			case "powershell":
				return cmd.Root().GenPowerShellCompletionWithDesc(cmd.OutOrStdout())
			}
			return nil
		},
	}
	return cmd
}
