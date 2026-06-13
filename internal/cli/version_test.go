package cli

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestVersionCommand(t *testing.T) {
	var out bytes.Buffer
	root := NewRootCmd()
	root.SetOut(&out)
	root.SetArgs([]string{"version"})
	require.NoError(t, root.Execute())
	require.Contains(t, out.String(), "hookfire")
}
