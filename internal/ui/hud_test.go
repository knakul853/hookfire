package ui

import (
	"bytes"
	"os"
	"regexp"
	"strings"
	"testing"
	"unicode/utf8"

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

func TestStripSchemeEdgeCases(t *testing.T) {
	require.Equal(t, "localhost:3000/webhook", stripScheme("http://localhost:3000/webhook"))
	require.Equal(t, "example.com/h?t=1", stripScheme("https://example.com/h?t=1"))
	require.Equal(t, "no-scheme/path", stripScheme("no-scheme/path"))
}

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

func TestHUDDryRunHeadersSorted(t *testing.T) {
	v := sampleView()
	v.DryRun = true
	v.Response = nil
	v.Request.Headers = map[string]string{"Z-Last": "1", "A-First": "2", "M-Mid": "3"}
	var buf bytes.Buffer
	require.NoError(t, RenderHUD(&buf, v))
	out := stripANSI(buf.String())
	require.Less(t, strings.Index(out, "A-First"), strings.Index(out, "M-Mid"))
	require.Less(t, strings.Index(out, "M-Mid"), strings.Index(out, "Z-Last"))
}

func TestTruncateIsRuneSafe(t *testing.T) {
	got := truncate(strings.Repeat("é", 100), 10)
	require.True(t, utf8.ValidString(got))
	require.LessOrEqual(t, utf8.RuneCountInString(got), 10)
}
