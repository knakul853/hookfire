package cli

import "github.com/spf13/cobra"

// rootDeps carries the shared flags resolved at the root level and passed down
// to subcommands via closure. ProvidersDir feeds buildCatalog in each command
// that needs a catalog (fire, list, etc.).
type rootDeps struct {
	ProvidersDir string
}

// NewRootCmd builds the hookfire command tree. main.go calls Execute on it.
func NewRootCmd() *cobra.Command {
	deps := &rootDeps{}
	root := &cobra.Command{
		Use:           "hookfire",
		Short:         "Fire correctly-signed synthetic webhook events at any URL",
		SilenceUsage:  true,
		SilenceErrors: true,
	}
	root.PersistentFlags().StringVar(&deps.ProvidersDir, "providers-dir", "", "extra providers directory (highest precedence)")
	root.AddCommand(newVersionCmd())
	return root
}

// Run executes the root command with args and returns a process exit code.
func Run(args []string) int {
	root := NewRootCmd()
	root.SetArgs(args)
	if err := root.Execute(); err != nil {
		return exitCodeFor(err)
	}
	return 0
}
