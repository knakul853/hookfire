package cli

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestParseSets(t *testing.T) {
	got, err := parseSets([]string{"a.b=1", "c=true"}, []string{"d=007"})
	require.NoError(t, err)
	require.Len(t, got, 3)
	require.Equal(t, "a.b", got[0].Path)
	require.False(t, got[0].ForceString)
	require.True(t, got[2].ForceString)

	_, err = parseSets([]string{"bad"}, nil)
	require.Error(t, err)
}

func TestParseHeaders(t *testing.T) {
	got, err := parseHeaders([]string{"X-A=1", "X-B=2"})
	require.NoError(t, err)
	require.Equal(t, "1", got["X-A"])
	_, err = parseHeaders([]string{"bad"})
	require.Error(t, err)
}
