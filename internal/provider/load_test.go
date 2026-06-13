package provider

import (
	"testing"
	"testing/fstest"

	"github.com/stretchr/testify/require"
)

func testFS() fstest.MapFS {
	return fstest.MapFS{
		"github/manifest.yaml":                   {Data: []byte(githubManifest)},
		"github/events/push.json":                {Data: []byte(`{"ref":"refs/heads/main"}`)},
		"github/events/pull_request.opened.json": {Data: []byte(`{"action":"opened"}`)},
	}
}

func TestLoadProviderDiscoversEvents(t *testing.T) {
	p, err := loadProvider(testFS(), "github")
	require.NoError(t, err)
	require.Equal(t, "github", p.Manifest.Name)
	require.Contains(t, p.Events, "push")
	require.Contains(t, p.Events, "pull_request.opened")
	require.JSONEq(t, `{"action":"opened"}`, string(p.Events["pull_request.opened"].Template))
}

func TestLoadProviderRejectsInvalidManifest(t *testing.T) {
	fsys := fstest.MapFS{"bad/manifest.yaml": {Data: []byte("signing:\n  scheme: magic\n")}}
	_, err := loadProvider(fsys, "bad")
	require.Error(t, err)
}

func TestLoadProviderMissingManifest(t *testing.T) {
	fsys := testFS()
	delete(fsys, "github/manifest.yaml")
	_, err := loadProvider(fsys, "github")
	require.Error(t, err)
}
