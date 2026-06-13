package cli

import (
	"errors"
	"fmt"

	"github.com/knakul853/hookfire/internal/provider"
	"github.com/knakul853/hookfire/internal/sign"
)

// errUsage marks a user-input/usage error (unknown target alias, missing URL,
// malformed --set/--header). SPEC §7 maps these to exit 2 alongside unknown
// provider/event.
var errUsage = errors.New("usage error")

// transportError marks a failure to reach the target (exit 4).
type transportError struct{ err error }

func (e *transportError) Error() string { return e.err.Error() }
func (e *transportError) Unwrap() error { return e.err }

// failResponseError marks a non-2xx response when --fail is set (exit 22).
type failResponseError struct{ status int }

func (e *failResponseError) Error() string { return fmt.Sprintf("target returned %d", e.status) }

// exitCodeFor maps an error to the documented process exit code (SPEC §7).
func exitCodeFor(err error) int {
	switch {
	case err == nil:
		return 0
	case errors.Is(err, provider.ErrUnknownProvider), errors.Is(err, provider.ErrUnknownEvent), errors.Is(err, errUsage):
		return 2
	case errors.Is(err, sign.ErrMissingSecret), errors.Is(err, sign.ErrUnsupportedScheme), errors.Is(err, sign.ErrInvalidConfig):
		return 3
	case isTransportError(err):
		return 4
	case isFailResponse(err):
		return 22
	default:
		return 1
	}
}

func isTransportError(err error) bool {
	var t *transportError
	return errors.As(err, &t)
}

func isFailResponse(err error) bool {
	var f *failResponseError
	return errors.As(err, &f)
}
