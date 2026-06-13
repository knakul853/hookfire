package cli

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestTriggerDryRunJSON(t *testing.T) {
	root := NewRootCmd()
	var out bytes.Buffer
	root.SetOut(&out)
	root.SetArgs([]string{"trigger", "github", "push", "--dry-run", "--json", "--url", "http://x/webhook", "--secret", "s"})
	require.NoError(t, root.Execute())

	var got map[string]any
	require.NoError(t, json.Unmarshal(out.Bytes(), &got))
	require.Equal(t, "github", got["provider"])
	require.Equal(t, "push", got["event"])
	require.Equal(t, true, got["dry_run"])
	require.Equal(t, true, got["signed"])
}

func TestTriggerUnknownProviderExit2(t *testing.T) {
	require.Equal(t, 2, Run([]string{"trigger", "gthub", "push", "--url", "http://x", "--secret", "s"}))
}

func TestTriggerFiresAtTestServer(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.NotEmpty(t, r.Header.Get("X-Hub-Signature-256"))
		w.WriteHeader(200)
		_, _ = w.Write([]byte(`{"ok":true}`))
	}))
	defer srv.Close()

	require.Equal(t, 0, Run([]string{"trigger", "github", "push", "--url", srv.URL, "--secret", "s", "--json"}))
}

func TestShowNeverFires(t *testing.T) {
	// show points at an unroutable URL; if it tried to send it would error, but
	// show is dry-run so it must succeed without touching the network.
	require.Equal(t, 0, Run([]string{"show", "github", "push", "--url", "http://127.0.0.1:1/x", "--secret", "s", "--json"}))
}

func TestDryRunNeedsNoURL(t *testing.T) {
	// show / --dry-run never hit the network, so a target URL is not required.
	require.Equal(t, 0, Run([]string{"show", "github", "push", "--secret", "s", "--json"}))
	require.Equal(t, 0, Run([]string{"trigger", "github", "push", "--dry-run", "--secret", "s", "--json"}))
}
