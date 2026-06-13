package ui

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestRenderDispatchAllModes(t *testing.T) {
	v := sampleView()

	var hud bytes.Buffer
	require.NoError(t, Render(&hud, ModeHUD, v))
	require.Contains(t, stripANSI(hud.String()), "github/pull_request.opened")

	var plain bytes.Buffer
	require.NoError(t, Render(&plain, ModePlain, v))
	require.Contains(t, plain.String(), "hookfire  github/pull_request.opened")

	var js bytes.Buffer
	require.NoError(t, Render(&js, ModeJSON, v))
	require.Contains(t, js.String(), `"provider": "github"`)
}

func TestStripSchemeEdgeCases(t *testing.T) {
	require.Equal(t, "localhost:3000/webhook", stripScheme("http://localhost:3000/webhook"))
	require.Equal(t, "example.com/h?t=1", stripScheme("https://example.com/h?t=1"))
	require.Equal(t, "no-scheme/path", stripScheme("no-scheme/path"))
}
