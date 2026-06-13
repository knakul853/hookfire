package sign

import (
	"strings"
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
	h, err := s.Sign(goldBody, Options{Secret: goldSecret, Timestamp: goldTS})
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
		Options{Secret: "It's a Secret to Everybody", Timestamp: goldTS})
	require.NoError(t, err)
	require.Equal(t,
		"sha256=757107ea0eb2509fc211221cce984b8a37570b6d7586c22c46f4379c8b043e17",
		h.Get("X-Hub-Signature-256"))
}

func TestHMACMissingSecret(t *testing.T) {
	s, err := New(githubConfig())
	require.NoError(t, err)
	_, err = s.Sign(goldBody, Options{Timestamp: goldTS})
	require.ErrorIs(t, err, ErrMissingSecret)
}

func TestHMACSlackGolden(t *testing.T) {
	cfg := SigningConfig{
		Scheme: "hmac", Algorithm: "sha256", Encoding: "hex",
		Basestring: "v0:{{timestamp}}:{{body}}", Output: "v0={{sig}}",
		Header:     "X-Slack-Signature",
		AuxHeaders: map[string]string{"X-Slack-Request-Timestamp": "{{timestamp}}"},
	}
	s, err := New(cfg)
	require.NoError(t, err)
	h, err := s.Sign(goldBody, Options{Secret: goldSecret, Timestamp: goldTS})
	require.NoError(t, err)
	require.Equal(t,
		"v0=473d71e21d4553ad4d6a2967b67acc609bad37b32502c07f06a00cfb310fd907",
		h.Get("X-Slack-Signature"))
	require.Equal(t, "1700000000", h.Get("X-Slack-Request-Timestamp"))
}

func TestHMACStripeGolden(t *testing.T) {
	cfg := SigningConfig{
		Scheme: "hmac", Algorithm: "sha256", Encoding: "hex",
		Basestring: "{{timestamp}}.{{body}}", Output: "t={{timestamp}},v1={{sig}}",
		Header: "Stripe-Signature",
	}
	s, err := New(cfg)
	require.NoError(t, err)
	h, err := s.Sign(goldBody, Options{Secret: goldSecret, Timestamp: goldTS})
	require.NoError(t, err)
	require.Equal(t,
		"t=1700000000,v1=2515c9ce1475bfae7728499a106d498a017d4d494a0f651800f25d06603e15ba",
		h.Get("Stripe-Signature"))
}

func TestHMACVercelGolden(t *testing.T) {
	cfg := SigningConfig{
		Scheme: "hmac", Algorithm: "sha1", Encoding: "hex",
		Basestring: "{{body}}", Output: "{{sig}}", Header: "x-vercel-signature",
	}
	s, err := New(cfg)
	require.NoError(t, err)
	h, err := s.Sign(goldBody, Options{Secret: goldSecret, Timestamp: goldTS})
	require.NoError(t, err)
	require.Equal(t, "8a09b3e005517381b23824cee6012c0771cbf01b", h.Get("x-vercel-signature"))
}

func TestHMACBase64Encoding(t *testing.T) {
	cfg := SigningConfig{
		Scheme: "hmac", Algorithm: "sha256", Encoding: "base64",
		Basestring: "{{body}}", Output: "{{sig}}", Header: "X-Sig",
	}
	s, err := New(cfg)
	require.NoError(t, err)
	h, err := s.Sign(goldBody, Options{Secret: goldSecret, Timestamp: goldTS})
	require.NoError(t, err)
	require.Equal(t, "ooRSC9bafvRbmGgC4HDThXbQqbmc81EPys4955Z22pE=", h.Get("X-Sig"))
}

func TestHMACDoesNotMutateBody(t *testing.T) {
	body := []byte(`{"hello":"world"}`)
	cp := append([]byte(nil), body...)
	s, err := New(githubConfig())
	require.NoError(t, err)
	_, err = s.Sign(body, Options{Secret: goldSecret, Timestamp: goldTS})
	require.NoError(t, err)
	require.Equal(t, cp, body)
}

func TestNewRejectsBadAlgorithm(t *testing.T) {
	_, err := New(SigningConfig{Scheme: "hmac", Algorithm: "md5", Encoding: "hex", Header: "X"})
	require.ErrorIs(t, err, ErrUnsupportedScheme)
}

func TestNewRejectsMissingHeader(t *testing.T) {
	_, err := New(SigningConfig{Scheme: "hmac", Algorithm: "sha256", Encoding: "hex"})
	require.ErrorIs(t, err, ErrInvalidConfig)
}

func TestSecretNeverInHeader(t *testing.T) {
	s, err := New(githubConfig())
	require.NoError(t, err)
	h, err := s.Sign(goldBody, Options{Secret: goldSecret, Timestamp: goldTS})
	require.NoError(t, err)
	for _, vals := range h {
		for _, v := range vals {
			require.False(t, strings.Contains(v, goldSecret), "secret leaked into header")
		}
	}
}
