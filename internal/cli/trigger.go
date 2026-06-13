package cli

import (
	"fmt"
	"log/slog"

	"github.com/knakul853/hookfire/internal/config"
	"github.com/knakul853/hookfire/internal/fire"
	"github.com/knakul853/hookfire/internal/sign"
	"github.com/knakul853/hookfire/internal/ui"
	"github.com/spf13/cobra"
)

func newTriggerCmd() *cobra.Command {
	var f commonFlags
	cmd := &cobra.Command{
		Use:   "trigger <provider> <event>",
		Short: "Render, sign, and fire a synthetic event at the target",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runTrigger(cmd, &f, args[0], args[1], false)
		},
	}
	bindCommonFlags(cmd, &f)
	return cmd
}

func newShowCmd() *cobra.Command {
	var f commonFlags
	cmd := &cobra.Command{
		Use:   "show <provider> <event>",
		Short: "Render and sign an event without sending it (dry-run)",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runTrigger(cmd, &f, args[0], args[1], true)
		},
	}
	bindCommonFlags(cmd, &f)
	return cmd
}

// runTrigger resolves config/secret, runs the pipeline, renders the view, and
// returns the pipeline error (mapped to an exit code by Run). show forces dryRun.
func runTrigger(cmd *cobra.Command, f *commonFlags, providerName, event string, forceDryRun bool) error {
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

	manifest, _, err := cat.Lookup(providerName, event)
	if err != nil {
		return suggestErr(cat, providerName, event, err)
	}

	if f.secret != "" {
		slog.Warn("--secret exposes the secret in shell history; prefer --secret-env")
	}
	if f.insecure {
		slog.Warn("--insecure disables TLS verification")
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

	sets, err := parseSets(f.sets, f.setStrings)
	if err != nil {
		return err
	}
	headers, err := parseHeaders(f.headers)
	if err != nil {
		return err
	}

	deps := pipelineDeps{Catalog: cat, Sender: fire.NewSender(fire.Options{Insecure: f.insecure})}
	view, perr := runPipeline(deps, pipelineInput{
		Provider:     providerName,
		Event:        event,
		URL:          url,
		Secret:       secret,
		Timestamp:    f.timestamp,
		Sets:         sets,
		ExtraHeaders: headers,
		NoSign:       f.noSign,
		DryRun:       f.dryRun || forceDryRun,
		Fail:         f.fail,
	})

	mode := ui.SelectMode(ui.DetectEnv(f.jsonOut))
	if rerr := ui.Render(cmd.OutOrStdout(), mode, view); rerr != nil {
		return rerr
	}
	return perr
}
