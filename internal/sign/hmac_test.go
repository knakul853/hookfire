package sign

import (
	"testing"

	"github.com/stretchr/testify/require"
)

const (
	goldSecret = "hookfire_test_secret"
	goldTS     = int64(1700000000)
)

var goldBody = []byte(`{"hello":"world"}`)

func githubConfig() SigningConfig {
	return SigningConfig{
		Scheme: "hmac", Algorithm: "sha256", Encoding: "hex",
		Basestring: "{{body}}", Output: "sha256={{sig}}",
		Header: "X-Hub-Signature-256",
	}
}

func TestHMACGithubGolden(t *testing.T) {
	s, err := New(githubConfig())
	require.NoError(t, err)
	h, err := s.Sign(goldBody, SignOptions{Secret: goldSecret, Timestamp: goldTS})
	require.NoError(t, err)
	require.Equal(t,
		"sha256=a284520bd6da7ef45b986802e070d38576d0a9b99cf3510fcace3de79676da91",
		h.Get("X-Hub-Signature-256"))
}

// Conformance anchor: GitHub's published canonical vector.
// https://docs.github.com/en/webhooks/using-webhooks/validating-webhook-deliveries
func TestHMACGithubPublishedCanonical(t *testing.T) {
	s, err := New(githubConfig())
	require.NoError(t, err)
	h, err := s.Sign([]byte("Hello, World!"),
		SignOptions{Secret: "It's a Secret to Everybody", Timestamp: goldTS})
	require.NoError(t, err)
	require.Equal(t,
		"sha256=757107ea0eb2509fc211221cce984b8a37570b6d7586c22c46f4379c8b043e17",
		h.Get("X-Hub-Signature-256"))
}

func TestHMACMissingSecret(t *testing.T) {
	s, err := New(githubConfig())
	require.NoError(t, err)
	_, err = s.Sign(goldBody, SignOptions{Timestamp: goldTS})
	require.ErrorIs(t, err, ErrMissingSecret)
}
