package cli

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNearest(t *testing.T) {
	got, ok := nearest("gthub", []string{"github", "slack", "stripe", "vercel"})
	require.True(t, ok)
	require.Equal(t, "github", got)

	_, ok = nearest("zzzzzzzz", []string{"github", "slack"})
	require.False(t, ok)
}
