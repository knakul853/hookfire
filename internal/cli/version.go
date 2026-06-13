package cli

import (
	"runtime/debug"

	"github.com/spf13/cobra"
)

// version is overridden at build time via -ldflags; "dev" when built plainly.
var version = "dev"

func newVersionCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Print the hookfire version",
		Run: func(cmd *cobra.Command, _ []string) {
			v := version
			if v == "dev" {
				if bi, ok := debug.ReadBuildInfo(); ok && bi.Main.Version != "" {
					v = bi.Main.Version
				}
			}
			cmd.Printf("hookfire %s\n", v)
		},
	}
}
