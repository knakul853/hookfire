package cli

import (
	"regexp"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNewUUIDFormat(t *testing.T) {
	re := regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-4[0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$`)
	require.Regexp(t, re, newUUID())
	require.NotEqual(t, newUUID(), newUUID())
}
