package config

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSecretPrecedence(t *testing.T) {
	t.Setenv("FLAG_ENV", "from_flag_env")
	t.Setenv("TARGET_ENV", "from_target_env")
	t.Setenv("MANIFEST_ENV", "from_manifest")
	r := EnvResolver{}

	got, src, err := ResolveSecret(r, ResolveInput{
		LiteralSecret: "literal", SecretEnvFlag: "FLAG_ENV",
		TargetSecretEnv: "TARGET_ENV", ManifestSource: "env:MANIFEST_ENV",
	})
	require.NoError(t, err)
	require.Equal(t, "literal", got.Reveal())
	require.Equal(t, SourceLiteral, src)

	got, src, err = ResolveSecret(r, ResolveInput{
		SecretEnvFlag: "FLAG_ENV", TargetSecretEnv: "TARGET_ENV", ManifestSource: "env:MANIFEST_ENV",
	})
	require.NoError(t, err)
	require.Equal(t, "from_flag_env", got.Reveal())
	require.Equal(t, SourceSecretEnvFlag, src)

	got, _, err = ResolveSecret(r, ResolveInput{TargetSecretEnv: "TARGET_ENV", ManifestSource: "env:MANIFEST_ENV"})
	require.NoError(t, err)
	require.Equal(t, "from_target_env", got.Reveal())

	got, _, err = ResolveSecret(r, ResolveInput{ManifestSource: "env:MANIFEST_ENV"})
	require.NoError(t, err)
	require.Equal(t, "from_manifest", got.Reveal())
}

func TestResolveSecretNoneAllowsEmpty(t *testing.T) {
	got, src, err := ResolveSecret(EnvResolver{}, ResolveInput{SchemeNone: true})
	require.NoError(t, err)
	require.Equal(t, "", got.Reveal())
	require.Equal(t, SourceNone, src)
}

func TestResolveSecretMissingErrors(t *testing.T) {
	_, _, err := ResolveSecret(EnvResolver{}, ResolveInput{})
	require.Error(t, err)
}
