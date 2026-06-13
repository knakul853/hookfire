package provider

import (
	"testing"

	"github.com/knakul853/hookfire/internal/sign"
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

func TestBuiltinGoldenEndToEnd(t *testing.T) {
	c, err := NewCatalog(Sources{Embedded: Builtin()})
	require.NoError(t, err)
	body := []byte(`{"hello":"world"}`)
	opts := sign.Options{Secret: "hookfire_test_secret", Timestamp: 1700000000}

	want := map[string]struct{ header, value string }{
		"github": {"X-Hub-Signature-256", "sha256=a284520bd6da7ef45b986802e070d38576d0a9b99cf3510fcace3de79676da91"},
		"slack":  {"X-Slack-Signature", "v0=473d71e21d4553ad4d6a2967b67acc609bad37b32502c07f06a00cfb310fd907"},
		"stripe": {"Stripe-Signature", "t=1700000000,v1=2515c9ce1475bfae7728499a106d498a017d4d494a0f651800f25d06603e15ba"},
		"vercel": {"x-vercel-signature", "8a09b3e005517381b23824cee6012c0771cbf01b"},
	}
	for prov, exp := range want {
		m, _, err := c.Lookup(prov, firstEvent(t, c, prov))
		require.NoError(t, err)
		s, err := sign.New(m.SigningConfig())
		require.NoError(t, err)
		h, err := s.Sign(body, opts)
		require.NoError(t, err)
		require.Equal(t, exp.value, h.Get(exp.header), "provider %s", prov)
	}
}
