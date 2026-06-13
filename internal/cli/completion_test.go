package cli

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/require"
)

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
