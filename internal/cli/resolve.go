package cli

import (
	"fmt"
	"log/slog"

	"github.com/knakul853/hookfire/internal/config"
	"github.com/knakul853/hookfire/internal/provider"
	"github.com/knakul853/hookfire/internal/sign"
)

// resolveTarget builds the catalog and resolves the config target alias and URL,
// applying the shared risky-flag warnings. It is used by trigger, show, and
// replay so they share one resolution path.
func resolveTarget(f *commonFlags) (provider.Catalog, config.Target, string, error) {
	cat, err := buildCatalog(f.providersDir, func(name, src string) {
		slog.Warn("filesystem provider shadows a built-in", "provider", name, "source", src)
	})
	if err != nil {
		return nil, config.Target{}, "", err
	}

	cfg := loadConfig()
	var tgt config.Target
	if f.target != "" {
		t, ok := cfg.Target(f.target)
		if !ok {
			return nil, config.Target{}, "", fmt.Errorf("%w: unknown target alias %q", errUsage, f.target)
		}
		tgt = t
	}

	url := f.url
	if url == "" {
		url = tgt.URL
	}
	if url == "" {
		return nil, config.Target{}, "", fmt.Errorf("%w: no target URL (pass --url or -t <alias> with a configured url)", errUsage)
	}

	if f.secret != "" {
		slog.Warn("--secret exposes the secret in shell history; prefer --secret-env")
	}
	if f.insecure {
		slog.Warn("--insecure disables TLS verification")
	}
	return cat, tgt, url, nil
}

// resolveSecret resolves the signing secret for manifest m per the precedence in
// SPEC §8 (--secret > --secret-env > target.secret_env > manifest source). It
// returns an empty secret when signing is skipped (--no-sign or scheme none); a
// resolution failure is reported as a missing-secret error (exit 3).
func resolveSecret(f *commonFlags, tgt config.Target, m provider.Manifest) (config.Secret, error) {
	if f.noSign || m.Signing.Scheme == "none" {
		return "", nil
	}
	s, src, err := config.ResolveSecret(config.EnvResolver{}, config.ResolveInput{
		LiteralSecret:   f.secret,
		SecretEnvFlag:   f.secretEnv,
		TargetSecretEnv: tgt.SecretEnv,
		ManifestSource:  m.Signing.SecretSource,
	})
	if err != nil {
		return "", fmt.Errorf("%w: %v", sign.ErrMissingSecret, err)
	}
	slog.Debug("resolved signing secret", "source", src)
	return s, nil
}
