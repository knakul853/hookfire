package provider

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestEventMergesHeaders(t *testing.T) {
	m, err := ParseManifest([]byte(githubManifest))
	require.NoError(t, err)
	ev := Event{
		Name:     "pull_request.opened",
		Template: []byte(`{"action":"opened"}`),
		manifest: &m,
	}
	hdrs := ev.Headers()
	require.Equal(t, "application/json", hdrs["Content-Type"])
	require.Equal(t, "GitHub-Hookshot/hookfire", hdrs["User-Agent"])
	require.Equal(t, "pull_request", hdrs["X-GitHub-Event"]) // per-event override
}
