package config

import "fmt"

// SecretSource records which input supplied the secret (for debug logging).
type SecretSource string

const (
	SourceLiteral       SecretSource = "--secret"
	SourceSecretEnvFlag SecretSource = "--secret-env"
	SourceTargetEnv     SecretSource = "target.secret_env"
	SourceManifest      SecretSource = "manifest.signing.secret_source"
	SourceNone          SecretSource = "none"
)

// ResolveInput holds every candidate secret source in precedence order.
type ResolveInput struct {
	LiteralSecret   string // --secret (highest)
	SecretEnvFlag   string // --secret-env NAME
	TargetSecretEnv string // target.secret_env NAME
	ManifestSource  string // manifest signing.secret_source ("env:NAME")
	SchemeNone      bool   // signing scheme is "none" → empty secret is valid
}

// ResolveSecret applies the precedence --secret → --secret-env → target.secret_env
// → manifest source. It returns the resolved Secret and which source it came
// from. When SchemeNone is set and no secret is supplied, it returns an empty
// secret without error.
func ResolveSecret(r SecretResolver, in ResolveInput) (Secret, SecretSource, error) {
	switch {
	case in.LiteralSecret != "":
		return Secret(in.LiteralSecret), SourceLiteral, nil
	case in.SecretEnvFlag != "":
		s, err := r.Resolve("env:" + in.SecretEnvFlag)
		return s, SourceSecretEnvFlag, err
	case in.TargetSecretEnv != "":
		s, err := r.Resolve("env:" + in.TargetSecretEnv)
		return s, SourceTargetEnv, err
	case in.ManifestSource != "":
		s, err := r.Resolve(in.ManifestSource)
		return s, SourceManifest, err
	case in.SchemeNone:
		return "", SourceNone, nil
	default:
		return "", SourceNone, fmt.Errorf("config: no secret available (set --secret, --secret-env, target.secret_env, or manifest secret_source)")
	}
}
