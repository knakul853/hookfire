package ui

import (
	"bytes"
	"os"
	"strings"
	"testing"
	"unicode/utf8"

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
	// A cut inside a multi-byte rune must not produce invalid UTF-8.
	got := truncate(strings.Repeat("é", 100), 10)
	require.True(t, utf8.ValidString(got))
	require.LessOrEqual(t, utf8.RuneCountInString(got), 10)
}

func TestPlainRendererStates(t *testing.T) {
	cases := map[string]struct {
		mutate func(*View)
		want   string
	}{
		"nosign": {func(v *View) { v.NoSign = true }, "skipped (--no-sign)"},
		"dryrun": {func(v *View) { v.DryRun = true; v.Response = nil }, "dry-run (not sent)"},
		"error":  {func(v *View) { v.Response = nil; v.Err = "connection refused" }, "connection refused"},
	}
	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			v := sampleView()
			c.mutate(&v)
			var buf bytes.Buffer
			require.NoError(t, RenderPlain(&buf, v))
			require.Contains(t, buf.String(), c.want)
		})
	}
}
