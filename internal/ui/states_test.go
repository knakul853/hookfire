package ui

import (
	"bytes"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestHUDNoSignGolden(t *testing.T) {
	v := sampleView()
	v.NoSign = true
	v.Signed = false
	var buf bytes.Buffer
	require.NoError(t, RenderHUD(&buf, v))
	want, err := os.ReadFile("../../testdata/ui/hud_nosign.txt")
	require.NoError(t, err)
	require.Equal(t, string(want), stripANSI(buf.String()))
}

func TestHUDDryRunGolden(t *testing.T) {
	v := sampleView()
	v.DryRun = true
	v.Response = nil
	var buf bytes.Buffer
	require.NoError(t, RenderHUD(&buf, v))
	want, err := os.ReadFile("../../testdata/ui/hud_dryrun.txt")
	require.NoError(t, err)
	require.Equal(t, string(want), stripANSI(buf.String()))
}

func TestHUDErrorGolden(t *testing.T) {
	v := sampleView()
	v.Response = nil
	v.Err = "connection refused"
	var buf bytes.Buffer
	require.NoError(t, RenderHUD(&buf, v))
	want, err := os.ReadFile("../../testdata/ui/hud_error.txt")
	require.NoError(t, err)
	require.Equal(t, string(want), stripANSI(buf.String()))
}
