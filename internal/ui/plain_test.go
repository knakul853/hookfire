package ui

import (
	"bytes"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestPlainSuccessGolden(t *testing.T) {
	var buf bytes.Buffer
	require.NoError(t, RenderPlain(&buf, sampleView()))
	want, err := os.ReadFile("../../testdata/ui/plain_success.txt")
	require.NoError(t, err)
	require.Equal(t, string(want), buf.String())
}
