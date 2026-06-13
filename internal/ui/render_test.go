package ui

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestRenderDispatch(t *testing.T) {
	var buf bytes.Buffer
	require.NoError(t, Render(&buf, ModeJSON, sampleView()))
	require.Contains(t, buf.String(), `"provider": "github"`)
}

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
