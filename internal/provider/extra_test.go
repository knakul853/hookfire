package provider

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestParseManifestRejectsBadYAML(t *testing.T) {
	_, err := ParseManifest([]byte("name: [unterminated\n  : :"))
	require.Error(t, err)
}

func TestLoadProviderMissingManifest(t *testing.T) {
	fsys := testFS()
	delete(fsys, "github/manifest.yaml")
	_, err := loadProvider(fsys, "github")
	require.Error(t, err)
}
