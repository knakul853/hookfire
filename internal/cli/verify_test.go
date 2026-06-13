package cli

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/knakul853/hookfire/internal/provider"
	"github.com/stretchr/testify/require"
)

func TestRunVerifyAllPass(t *testing.T) {
	var buf bytes.Buffer
	err := runVerify(&buf, []string{"github", "slack"}, func(_ string) (string, error) {
		return "sha256=deadbeef", nil
	})
	require.NoError(t, err)
	require.Contains(t, buf.String(), "✓ github")
	require.Contains(t, buf.String(), "✓ slack")
}

func TestRunVerifyReportsFailures(t *testing.T) {
	var buf bytes.Buffer
	err := runVerify(&buf, []string{"good", "bad"}, func(name string) (string, error) {
		if name == "bad" {
			return "", errors.New("boom")
		}
		return "sig", nil
	})
	require.Error(t, err)
	require.Contains(t, buf.String(), "✗ bad: boom")
}

func TestVerifyCommandOnBuiltins(t *testing.T) {
	root := NewRootCmd()
	var out bytes.Buffer
	root.SetOut(&out)
	root.SetArgs([]string{"verify"})
	require.NoError(t, root.Execute())
	for _, p := range []string{"github", "slack", "stripe", "vercel"} {
		require.Contains(t, out.String(), "✓ "+p)
	}
}

func TestVerifyCommandOnBadProviderDir(t *testing.T) {
	dir := t.TempDir()
	require.NoError(t, writeBadProvider(dir))
	require.NotEqual(t, 0, Run([]string{"verify", dir}))
}

func writeBadProvider(dir string) error {
	provDir := filepath.Join(dir, "bad")
	if err := os.MkdirAll(provDir, 0o755); err != nil {
		return err
	}
	manifest := []byte("name: bad\nsigning:\n  scheme: magic\n")
	return os.WriteFile(filepath.Join(provDir, "manifest.yaml"), manifest, 0o644)
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
