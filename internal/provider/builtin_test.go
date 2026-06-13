package provider

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestBuiltinProvidersLoadAndValidate(t *testing.T) {
	c, err := NewCatalog(Sources{Embedded: Builtin()})
	require.NoError(t, err)
	require.ElementsMatch(t, []string{"github", "slack", "stripe", "vercel"}, c.List())
}

func TestBuiltinSigningMatchesGoldenTable(t *testing.T) {
	c, err := NewCatalog(Sources{Embedded: Builtin()})
	require.NoError(t, err)
	cases := map[string]string{
		"github": "X-Hub-Signature-256",
		"slack":  "X-Slack-Signature",
		"stripe": "Stripe-Signature",
		"vercel": "x-vercel-signature",
	}
	for prov, header := range cases {
		m, _, err := c.Lookup(prov, firstEvent(t, c, prov))
		require.NoError(t, err)
		require.Equal(t, header, m.Signing.Header, "provider %s header", prov)
	}
}

func firstEvent(t *testing.T, c Catalog, prov string) string {
	t.Helper()
	evs, err := c.Events(prov)
	require.NoError(t, err)
	require.NotEmpty(t, evs)
	return evs[0]
}
