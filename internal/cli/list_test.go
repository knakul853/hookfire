package cli

import (
	"bytes"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestListProviders(t *testing.T) {
	root := NewRootCmd()
	var out bytes.Buffer
	root.SetOut(&out)
	root.SetArgs([]string{"list", "--json"})
	require.NoError(t, root.Execute())
	var got []string
	require.NoError(t, json.Unmarshal(out.Bytes(), &got))
	require.Subset(t, got, []string{"github", "slack", "stripe", "vercel"})
}

func TestListEvents(t *testing.T) {
	root := NewRootCmd()
	var out bytes.Buffer
	root.SetOut(&out)
	root.SetArgs([]string{"list", "github", "--json"})
	require.NoError(t, root.Execute())
	var got []string
	require.NoError(t, json.Unmarshal(out.Bytes(), &got))
	require.Contains(t, got, "push")
}

func TestPrintListWritesToGivenWriter(t *testing.T) {
	var buf bytes.Buffer
	require.NoError(t, printList(&buf, []string{"a", "b"}, false))
	require.Equal(t, "a\nb\n", buf.String())

	buf.Reset()
	require.NoError(t, printList(&buf, []string{"a", "b"}, true))
	require.JSONEq(t, `["a","b"]`, buf.String())
}
