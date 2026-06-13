package cli

import (
	"testing"

	"github.com/knakul853/hookfire/internal/config"
	"github.com/knakul853/hookfire/internal/provider"
	"github.com/knakul853/hookfire/internal/sign"
	"github.com/stretchr/testify/require"
)

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
