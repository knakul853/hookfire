package cli

import (
	"bytes"
	"context"
	"errors"
	"log/slog"
	"testing"

	"github.com/knakul853/hookfire/internal/provider"
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

func TestCompletionEmitsScript(t *testing.T) {
	for _, shell := range []string{"bash", "zsh", "fish", "powershell"} {
		root := NewRootCmd()
		var out bytes.Buffer
		root.SetOut(&out)
		root.SetArgs([]string{"completion", shell})
		require.NoError(t, root.Execute(), shell)
		require.NotEmpty(t, out.String(), shell)
	}
}
