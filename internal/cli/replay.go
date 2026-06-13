package cli

import (
	"fmt"
	"log/slog"
	"os"

	"github.com/knakul853/hookfire/internal/config"
	"github.com/knakul853/hookfire/internal/fire"
	"github.com/knakul853/hookfire/internal/provider"
	"github.com/knakul853/hookfire/internal/sign"
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
	cat, err := buildCatalog(f.providersDir, func(name, src string) {
		slog.Warn("filesystem provider shadows a built-in", "provider", name, "source", src)
	})
	if err != nil {
		return err
	}

	cfg := loadConfig()
	var tgt config.Target
	if f.target != "" {
		t, ok := cfg.Target(f.target)
		if !ok {
			return fmt.Errorf("unknown target alias %q", f.target)
		}
		tgt = t
	}
	url := f.url
	if url == "" {
		url = tgt.URL
	}
	if url == "" {
		return fmt.Errorf("no target URL: pass --url or -t <alias> with a configured url")
	}

	manifest, err := cat.Manifest(providerName)
	if err != nil {
		return suggestErr(cat, providerName, "", err)
	}

	body, err := os.ReadFile(file)
	if err != nil {
		return fmt.Errorf("cli: read replay file: %w", err)
	}

	if f.secret != "" {
		slog.Warn("--secret exposes the secret in shell history; prefer --secret-env")
	}
	var secret config.Secret
	if !f.noSign && manifest.Signing.Scheme != "none" {
		s, src, rerr := config.ResolveSecret(config.EnvResolver{}, config.ResolveInput{
			LiteralSecret:   f.secret,
			SecretEnvFlag:   f.secretEnv,
			TargetSecretEnv: tgt.SecretEnv,
			ManifestSource:  manifest.Signing.SecretSource,
		})
		if rerr != nil {
			return fmt.Errorf("%w: %v", sign.ErrMissingSecret, rerr)
		}
		secret = s
		slog.Debug("resolved signing secret", "source", src)
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
