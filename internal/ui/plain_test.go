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
