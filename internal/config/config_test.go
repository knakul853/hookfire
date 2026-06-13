package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

const sampleTOML = `
[targets.local]
url = "http://localhost:3000/webhook"
provider = "github"
secret_env = "GITHUB_WEBHOOK_SECRET"

[targets.staging]
url = "https://staging.example.com/hooks/github"
secret_env = "STAGING_GH_SECRET"
`

func TestLoadConfigAndTargets(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.toml")
	require.NoError(t, os.WriteFile(path, []byte(sampleTOML), 0o600))
	cfg, err := Load(path)
	require.NoError(t, err)
	local, ok := cfg.Target("local")
	require.True(t, ok)
	require.Equal(t, "http://localhost:3000/webhook", local.URL)
	require.Equal(t, "github", local.Provider)
	require.Equal(t, "GITHUB_WEBHOOK_SECRET", local.SecretEnv)
	_, ok = cfg.Target("nope")
	require.False(t, ok)
}

func TestLoadMissingFileIsEmptyConfig(t *testing.T) {
	cfg, err := Load(filepath.Join(t.TempDir(), "absent.toml"))
	require.NoError(t, err)
	_, ok := cfg.Target("local")
	require.False(t, ok)
}

func TestDefaultPathHonorsXDG(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", "/tmp/xyz")
	p, err := DefaultPath()
	require.NoError(t, err)
	require.Equal(t, filepath.Join("/tmp/xyz", "hookfire", "config.toml"), p)
}

func TestDefaultPathFallsBackToHome(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", "")
	p, err := DefaultPath()
	require.NoError(t, err)
	home, herr := os.UserHomeDir()
	require.NoError(t, herr)
	require.Equal(t, filepath.Join(home, ".config", "hookfire", "config.toml"), p)
}

func TestLoadRejectsBadTOML(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, "config.toml")
	require.NoError(t, os.WriteFile(p, []byte("this = = not ]["), 0o600))
	_, err := Load(p)
	require.Error(t, err)
}
