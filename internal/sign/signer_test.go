package sign

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNewRejectsUnknownScheme(t *testing.T) {
	_, err := New(SigningConfig{Scheme: "magic"})
	require.ErrorIs(t, err, ErrUnsupportedScheme)
}

func TestNewReservesJWT(t *testing.T) {
	_, err := New(SigningConfig{Scheme: "jwt"})
	require.ErrorIs(t, err, ErrUnsupportedScheme)
}

func TestNewNoneNeedsNoSecret(t *testing.T) {
	s, err := New(SigningConfig{Scheme: "none"})
	require.NoError(t, err)
	h, err := s.Sign([]byte("x"), SignOptions{})
	require.NoError(t, err)
	require.Empty(t, h)
}

func TestErrUnsupportedSchemeIsSentinel(t *testing.T) {
	require.True(t, errors.Is(ErrUnsupportedScheme, ErrUnsupportedScheme))
}
