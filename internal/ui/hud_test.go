package ui

import (
	"bytes"
	"os"
	"regexp"
	"testing"

	"github.com/stretchr/testify/require"
)

var ansiRE = regexp.MustCompile(`\x1b\[[0-9;]*m`)

func stripANSI(s string) string {
	return ansiRE.ReplaceAllString(s, "")
}

func TestHUDSuccessGoldenStripped(t *testing.T) {
	v := sampleView()
	v.Overrides = 3
	var buf bytes.Buffer
	require.NoError(t, RenderHUD(&buf, v))
	want, err := os.ReadFile("../../testdata/ui/hud_success.txt")
	require.NoError(t, err)
	require.Equal(t, string(want), stripANSI(buf.String()))
}
