package config

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSecretRedactsEverywhere(t *testing.T) {
	s := Secret("supersecret")
	require.Equal(t, "***", s.String())
	require.Equal(t, "***", fmt.Sprintf("%v", s))
	require.Equal(t, "***", fmt.Sprintf("%s", s)) //nolint:staticcheck // intentional: verifies %s triggers String()
	require.NotContains(t, fmt.Sprintf("%#v", s), "supersecret")
	require.NotContains(t, fmt.Sprintf("%#v", struct{ S Secret }{s}), "supersecret")
	b, err := json.Marshal(struct{ S Secret }{s})
	require.NoError(t, err)
	require.JSONEq(t, `{"S":"***"}`, string(b))
	require.Equal(t, "supersecret", s.Reveal())
}

func TestEnvResolver(t *testing.T) {
	t.Setenv("HF_TEST_SECRET", "fromenv")
	r := EnvResolver{}
	got, err := r.Resolve("env:HF_TEST_SECRET")
	require.NoError(t, err)
	require.Equal(t, "fromenv", got.Reveal())
}

func TestEnvResolverMissing(t *testing.T) {
	r := EnvResolver{}
	_, err := r.Resolve("env:HF_DEFINITELY_UNSET")
	require.Error(t, err)
}
