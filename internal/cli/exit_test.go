package cli

import (
	"errors"
	"testing"

	"github.com/knakul853/hookfire/internal/provider"
	"github.com/knakul853/hookfire/internal/sign"
	"github.com/stretchr/testify/require"
)

func TestExitCodeFor(t *testing.T) {
	require.Equal(t, 0, exitCodeFor(nil))
	require.Equal(t, 2, exitCodeFor(provider.ErrUnknownProvider))
	require.Equal(t, 2, exitCodeFor(provider.ErrUnknownEvent))
	require.Equal(t, 3, exitCodeFor(sign.ErrMissingSecret))
	require.Equal(t, 3, exitCodeFor(sign.ErrUnsupportedScheme))
	require.Equal(t, 3, exitCodeFor(sign.ErrInvalidConfig))
	require.Equal(t, 4, exitCodeFor(&transportError{errors.New("refused")}))
	require.Equal(t, 22, exitCodeFor(&failResponseError{status: 500}))
	require.Equal(t, 1, exitCodeFor(errors.New("other")))
}
