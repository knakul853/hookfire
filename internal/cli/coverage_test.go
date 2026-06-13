package cli

import (
	"context"
	"errors"
	"log/slog"
	"testing"

	"github.com/knakul853/hookfire/internal/config"
	"github.com/knakul853/hookfire/internal/provider"
	"github.com/knakul853/hookfire/internal/sign"
	"github.com/stretchr/testify/require"
)

func builtinCat(t *testing.T) provider.Catalog {
	t.Helper()
	c, err := buildCatalog("", func(string, string) {})
	require.NoError(t, err)
	return c
}

func TestSuggestErr(t *testing.T) {
	cat := builtinCat(t)

	err := suggestErr(cat, "gthub", "push", provider.ErrUnknownProvider)
	require.ErrorIs(t, err, provider.ErrUnknownProvider)
	require.Contains(t, err.Error(), "github")

	err = suggestErr(cat, "github", "psh", provider.ErrUnknownEvent)
	require.ErrorIs(t, err, provider.ErrUnknownEvent)
	require.Contains(t, err.Error(), "push")

	err = suggestErr(cat, "zzzzzzzz", "x", provider.ErrUnknownProvider)
	require.NotContains(t, err.Error(), "did you mean")

	other := errors.New("boom")
	require.Equal(t, other, suggestErr(cat, "x", "y", other))
}

func TestResolveSecretSkipsWhenNotSigning(t *testing.T) {
	none, err := provider.ParseManifest([]byte("name: x\nsigning: { scheme: none }\n"))
	require.NoError(t, err)
	s, err := resolveSecret(&commonFlags{}, config.Target{}, none)
	require.NoError(t, err)
	require.Equal(t, "", s.Reveal())

	gh, err := provider.ParseManifest([]byte(githubManifestYAML))
	require.NoError(t, err)
	s, err = resolveSecret(&commonFlags{noSign: true}, config.Target{}, gh)
	require.NoError(t, err)
	require.Equal(t, "", s.Reveal())
}

func TestResolveSecretLiteralAndMissing(t *testing.T) {
	gh, err := provider.ParseManifest([]byte(githubManifestYAML))
	require.NoError(t, err)

	s, err := resolveSecret(&commonFlags{secret: "abc"}, config.Target{}, gh)
	require.NoError(t, err)
	require.Equal(t, "abc", s.Reveal())

	_, err = resolveSecret(&commonFlags{}, config.Target{}, gh)
	require.ErrorIs(t, err, sign.ErrMissingSecret)
}

func TestResolveTargetURLRules(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())

	_, _, _, err := resolveTarget(&commonFlags{}, true)
	require.ErrorIs(t, err, errUsage)

	_, _, url, err := resolveTarget(&commonFlags{}, false)
	require.NoError(t, err)
	require.Equal(t, "", url)

	_, _, _, err = resolveTarget(&commonFlags{target: "nope", url: "http://x"}, true)
	require.ErrorIs(t, err, errUsage)
}

func TestVerifyProviderHMACAndNone(t *testing.T) {
	gh, err := provider.ParseManifest([]byte(githubManifestYAML))
	require.NoError(t, err)
	sig, err := verifyProvider(gh)
	require.NoError(t, err)
	require.NotEmpty(t, sig)

	none, err := provider.ParseManifest([]byte("name: n\nsigning: { scheme: none }\n"))
	require.NoError(t, err)
	sig, err = verifyProvider(none)
	require.NoError(t, err)
	require.Equal(t, "", sig)
}

func TestConfigureLoggingLevels(t *testing.T) {
	ctx := context.Background()
	configureLogging(0)
	require.False(t, slog.Default().Enabled(ctx, slog.LevelInfo))
	configureLogging(1)
	require.True(t, slog.Default().Enabled(ctx, slog.LevelInfo))
	require.False(t, slog.Default().Enabled(ctx, slog.LevelDebug))
	configureLogging(2)
	require.True(t, slog.Default().Enabled(ctx, slog.LevelDebug))
}
