package cli

import (
	"fmt"
	"strings"

	"github.com/knakul853/hookfire/internal/render"
	"github.com/spf13/cobra"
)

// commonFlags are the flags shared by trigger, show, and replay.
type commonFlags struct {
	target       string
	url          string
	secret       string
	secretEnv    string
	sets         []string
	setStrings   []string
	headers      []string
	timestamp    int64
	noSign       bool
	dryRun       bool
	jsonOut      bool
	fail         bool
	insecure     bool
	providersDir string
}

func bindCommonFlags(cmd *cobra.Command, f *commonFlags) {
	fl := cmd.Flags()
	fl.StringVarP(&f.target, "target", "t", "", "Resolve url + secret (+ provider) from a config target alias")
	fl.StringVar(&f.url, "url", "", "Explicit target URL (overrides the target's url)")
	fl.StringVar(&f.secret, "secret", "", "Literal signing secret (discouraged; prefer --secret-env)")
	fl.StringVar(&f.secretEnv, "secret-env", "", "Read the signing secret from this environment variable")
	fl.StringArrayVar(&f.sets, "set", nil, "Override a payload field by dot-path (path=value); repeatable")
	fl.StringArrayVar(&f.setStrings, "set-string", nil, "Like --set but force a string value; repeatable")
	fl.StringArrayVar(&f.headers, "header", nil, "Add/override a request header (k=v); repeatable")
	fl.Int64Var(&f.timestamp, "timestamp", 0, "Pin the signing timestamp (unix seconds)")
	fl.BoolVar(&f.noSign, "no-sign", false, "Send unsigned (for testing rejection paths)")
	fl.BoolVar(&f.dryRun, "dry-run", false, "Render + sign, print, never touch the network")
	fl.BoolVar(&f.jsonOut, "json", false, "Machine-readable output (implies no ANSI)")
	fl.BoolVar(&f.fail, "fail", false, "Exit non-zero on a non-2xx target response")
	fl.BoolVar(&f.insecure, "insecure", false, "Skip TLS verification")
	fl.StringVar(&f.providersDir, "providers-dir", "", "Extra providers directory (highest precedence)")
}

func parseSets(sets, setStrings []string) ([]render.Set, error) {
	out := make([]render.Set, 0, len(sets)+len(setStrings))
	for _, s := range sets {
		k, v, ok := strings.Cut(s, "=")
		if !ok {
			return nil, fmt.Errorf("--set %q must be path=value", s)
		}
		out = append(out, render.Set{Path: k, Value: v})
	}
	for _, s := range setStrings {
		k, v, ok := strings.Cut(s, "=")
		if !ok {
			return nil, fmt.Errorf("--set-string %q must be path=value", s)
		}
		out = append(out, render.Set{Path: k, Value: v, ForceString: true})
	}
	return out, nil
}

func parseHeaders(hs []string) (map[string]string, error) {
	out := make(map[string]string, len(hs))
	for _, h := range hs {
		k, v, ok := strings.Cut(h, "=")
		if !ok {
			return nil, fmt.Errorf("--header %q must be k=v", h)
		}
		out[k] = v
	}
	return out, nil
}
