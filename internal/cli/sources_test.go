package cli

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestProviderDirPathsOrder(t *testing.T) {
	xdg := t.TempDir()
	cwd := t.TempDir()
	flag := t.TempDir()
	require.NoError(t, os.MkdirAll(filepath.Join(xdg, "hookfire", "providers"), 0o755))
	require.NoError(t, os.MkdirAll(filepath.Join(cwd, "providers"), 0o755))
	// flag dir itself is the providers dir

	paths := providerDirPaths(xdg, cwd, flag)
	require.Equal(t, []string{
		filepath.Join(xdg, "hookfire", "providers"),
		filepath.Join(cwd, "providers"),
		flag,
	}, paths)
}

func TestProviderDirPathsSkipsMissing(t *testing.T) {
	paths := providerDirPaths("/nonexistent", "/nonexistent", "")
	require.Empty(t, paths)
}
