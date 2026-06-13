package render

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestVarsAreDeterministicWhenInjected(t *testing.T) {
	v := Vars{Timestamp: 1700000000, UUID: "fixed-uuid", Event: "push", Now: "2023-11-14T22:13:20Z"}
	m := v.Map()
	require.Equal(t, "1700000000", m["timestamp"])
	require.Equal(t, "fixed-uuid", m["uuid"])
	require.Equal(t, "push", m["event"])
	require.Equal(t, "2023-11-14T22:13:20Z", m["now"])
	require.Equal(t, "2023-11-14T22:13:20Z", m["iso8601"])
}
