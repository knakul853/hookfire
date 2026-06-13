package cli

import "github.com/spf13/cobra"

// NewRootCmd builds the hookfire command tree. main.go calls Execute on it.
func NewRootCmd() *cobra.Command {
	root := &cobra.Command{
		Use:           "hookfire",
		Short:         "Fire correctly-signed synthetic webhook events at any URL",
		SilenceUsage:  true,
		SilenceErrors: true,
	}
	root.AddCommand(newVersionCmd())
	return root
}

// Run executes the root command with args and returns a process exit code.
// The full exit-code mapping arrives in a later milestone; this is the stub.
func Run(args []string) int {
	root := NewRootCmd()
	root.SetArgs(args)
	if err := root.Execute(); err != nil {
		return 1
	}
	return 0
}
