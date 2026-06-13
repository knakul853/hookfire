package cli

import (
	"encoding/json"
	"fmt"
	"io"

	"github.com/spf13/cobra"
)

func newListCmd() *cobra.Command {
	var providersDir string
	var jsonOut bool
	cmd := &cobra.Command{
		Use:   "list [provider]",
		Short: "List providers, or one provider's events",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cat, err := buildCatalog(providersDir, func(string, string) {})
			if err != nil {
				return err
			}
			if len(args) == 0 {
				return printList(cmd.OutOrStdout(), cat.List(), jsonOut)
			}
			evs, err := cat.Events(args[0])
			if err != nil {
				return suggestErr(cat, args[0], "", err)
			}
			return printList(cmd.OutOrStdout(), evs, jsonOut)
		},
	}
	cmd.Flags().StringVar(&providersDir, "providers-dir", "", "Extra providers directory (highest precedence)")
	cmd.Flags().BoolVar(&jsonOut, "json", false, "Machine-readable output")
	return cmd
}

// printList writes items to w. It writes to the explicit stdout writer (not
// cobra's Print family, which targets stderr) so command output stays on stdout.
func printList(w io.Writer, items []string, jsonOut bool) error {
	if jsonOut {
		b, err := json.Marshal(items)
		if err != nil {
			return fmt.Errorf("cli: marshal list: %w", err)
		}
		_, err = fmt.Fprintln(w, string(b))
		return err
	}
	for _, it := range items {
		if _, err := fmt.Fprintln(w, it); err != nil {
			return err
		}
	}
	return nil
}
