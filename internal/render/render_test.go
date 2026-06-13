package render

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestRenderInterpolatesAndCanonicalizes(t *testing.T) {
	tmpl := []byte(`{"event":"{{event}}","ts":{{timestamp}},"id":"{{uuid}}"}`)
	v := Vars{Timestamp: 1700000000, UUID: "abc", Event: "push", Now: "2023-11-14T22:13:20Z"}
	out, err := Render(tmpl, v, nil)
	require.NoError(t, err)
	require.JSONEq(t, `{"event":"push","ts":1700000000,"id":"abc"}`, string(out))
}

func TestRenderRejectsInvalidJSONAfterInterpolation(t *testing.T) {
	_, err := Render([]byte(`{"x": }`), Vars{}, nil)
	require.Error(t, err)
}

func TestRenderPreservesLargeIntegers(t *testing.T) {
	out, err := Render([]byte(`{"id":1234567890123456789}`), Vars{}, nil)
	require.NoError(t, err)
	require.Equal(t, `{"id":1234567890123456789}`, string(out))
}

func TestRenderUnknownVarStaysLiteral(t *testing.T) {
	out, err := Render([]byte(`{"x":"{{unknown}}"}`), Vars{}, nil)
	require.NoError(t, err)
	require.JSONEq(t, `{"x":"{{unknown}}"}`, string(out))
}
