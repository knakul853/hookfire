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
