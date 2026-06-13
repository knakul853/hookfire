package sign

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

// The secret must never appear in the produced header value.
func TestSecretNeverInHeader(t *testing.T) {
	s, _ := New(githubConfig())
	h, err := s.Sign(goldBody, SignOptions{Secret: goldSecret, Timestamp: goldTS})
	require.NoError(t, err)
	for _, vals := range h {
		for _, v := range vals {
			require.False(t, strings.Contains(v, goldSecret), "secret leaked into header")
		}
	}
}
