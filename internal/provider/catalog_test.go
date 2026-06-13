package provider

import (
	"io/fs"
	"testing"
	"testing/fstest"

	"github.com/stretchr/testify/require"
)

func TestCatalogLookupAndList(t *testing.T) {
	c, err := NewCatalog(Sources{Embedded: testFS()})
	require.NoError(t, err)
	m, ev, err := c.Lookup("github", "push")
	require.NoError(t, err)
	require.Equal(t, "github", m.Name)
	require.Equal(t, "push", ev.Name)
	require.ElementsMatch(t, []string{"github"}, c.List())
}

func TestCatalogUnknownProvider(t *testing.T) {
	c, _ := NewCatalog(Sources{Embedded: testFS()})
	_, _, err := c.Lookup("gitlab", "push")
	require.ErrorIs(t, err, ErrUnknownProvider)
}

func TestCatalogUnknownEvent(t *testing.T) {
	c, _ := NewCatalog(Sources{Embedded: testFS()})
	_, _, err := c.Lookup("github", "nope")
	require.ErrorIs(t, err, ErrUnknownEvent)
}

func TestCatalogFilesystemOverridesEmbeddedWithWarn(t *testing.T) {
	fsOverride := fstest.MapFS{
		"github/manifest.yaml":    {Data: []byte(githubManifest)},
		"github/events/push.json": {Data: []byte(`{"ref":"refs/heads/override"}`)},
	}
	var warned []string
	c, err := NewCatalog(Sources{
		Embedded: testFS(),
		Dirs:     []fs.FS{fsOverride},
		OnShadow: func(name, _ string) { warned = append(warned, name) },
	})
	require.NoError(t, err)
	_, ev, _ := c.Lookup("github", "push")
	require.JSONEq(t, `{"ref":"refs/heads/override"}`, string(ev.Template))
	require.Equal(t, []string{"github"}, warned)
}
