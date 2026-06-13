package sign

import (
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
	h, err := s.Sign([]byte("x"), Options{})
	require.NoError(t, err)
	require.Empty(t, h)
}

func TestNewErrorsCarryContext(t *testing.T) {
	_, err := New(SigningConfig{Scheme: "hmac", Algorithm: "md5", Encoding: "hex", Header: "X"})
	require.ErrorIs(t, err, ErrUnsupportedScheme)
	require.Contains(t, err.Error(), "md5")
}
