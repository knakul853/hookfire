package cli

import (
	"github.com/knakul853/hookfire/internal/fire"
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
	cat, tgt, url, err := resolveTarget(f)
	if err != nil {
		return err
	}

	manifest, _, err := cat.Lookup(providerName, event)
	if err != nil {
		return suggestErr(cat, providerName, event, err)
	}

	secret, err := resolveSecret(f, tgt, manifest)
	if err != nil {
		return err
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
