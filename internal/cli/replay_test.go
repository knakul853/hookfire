package cli

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestReplaySignsFileBodyVerbatim(t *testing.T) {
	const payload = `{"replayed":true}`
	dir := t.TempDir()
	file := filepath.Join(dir, "payload.json")
	require.NoError(t, os.WriteFile(file, []byte(payload), 0o600))

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body := make([]byte, r.ContentLength)
		_, _ = r.Body.Read(body)
		require.Equal(t, payload, string(body))
		require.NotEmpty(t, r.Header.Get("X-Hub-Signature-256"))
		w.WriteHeader(200)
	}))
	defer srv.Close()

	require.Equal(t, 0, Run([]string{"replay", file, "--provider", "github", "--url", srv.URL, "--secret", "s"}))
}

func TestReplayMissingProviderFlag(t *testing.T) {
	require.NotEqual(t, 0, Run([]string{"replay", "/tmp/whatever.json", "--url", "http://x"}))
}
