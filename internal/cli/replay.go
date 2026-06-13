package cli

import (
	"fmt"
	"os"

	"github.com/knakul853/hookfire/internal/fire"
	"github.com/knakul853/hookfire/internal/provider"
	"github.com/knakul853/hookfire/internal/ui"
	"github.com/spf13/cobra"
)

func newReplayCmd() *cobra.Command {
	var f commonFlags
	var providerName string
	cmd := &cobra.Command{
		Use:   "replay <file.json>",
		Short: "Re-sign and fire a saved JSON payload",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runReplay(cmd, &f, providerName, args[0])
		},
	}
	bindCommonFlags(cmd, &f)
	cmd.Flags().StringVar(&providerName, "provider", "", "Provider whose signing scheme to use (required)")
	if err := cmd.MarkFlagRequired("provider"); err != nil {
		panic(err)
	}
	return cmd
}

// runReplay reads a saved payload verbatim (no rendering, no --set), signs it
// with the named provider's scheme, and fires it. The body is sent byte-for-byte.
func runReplay(cmd *cobra.Command, f *commonFlags, providerName, file string) error {
	cat, tgt, url, err := resolveTarget(f, !f.dryRun)
	if err != nil {
		return err
	}

	manifest, err := cat.Manifest(providerName)
	if err != nil {
		return suggestErr(cat, providerName, "", err)
	}

	body, err := os.ReadFile(file)
	if err != nil {
		return fmt.Errorf("cli: read replay file: %w", err)
	}

	secret, err := resolveSecret(f, tgt, manifest)
	if err != nil {
		return err
	}
	headers, err := parseHeaders(f.headers)
	if err != nil {
		return err
	}

	deps := pipelineDeps{Catalog: cat, Sender: fire.NewSender(fire.Options{Insecure: f.insecure})}
	view, perr := fireResult(deps, fireParams{
		manifest:     manifest,
		provider:     providerName,
		url:          url,
		body:         body,
		secret:       secret,
		timestamp:    f.timestamp,
		baseHeaders:  replayHeaders(manifest),
		extraHeaders: headers,
		noSign:       f.noSign,
		dryRun:       f.dryRun,
		fail:         f.fail,
	})
	mode := ui.SelectMode(ui.DetectEnv(f.jsonOut))
	if rerr := ui.Render(cmd.OutOrStdout(), mode, view); rerr != nil {
		return rerr
	}
	return perr
}

// replayHeaders builds the transport-only headers for a replay (no per-event
// headers, since replay has no event): Content-Type plus manifest transport headers.
func replayHeaders(m provider.Manifest) map[string]string {
	out := map[string]string{}
	if m.Transport.ContentType != "" {
		out["Content-Type"] = m.Transport.ContentType
	}
	for k, v := range m.Transport.Headers {
		out[k] = v
	}
	return out
}
