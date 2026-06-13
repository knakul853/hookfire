package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

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

func TestEnvResolverRejectsNonEnvSource(t *testing.T) {
	_, err := EnvResolver{}.Resolve("vault:secret/foo")
	require.Error(t, err)
}
