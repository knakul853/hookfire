package config

import (
	"fmt"
	"os"
	"strings"
)

// Secret is a signing secret that redacts itself everywhere except Reveal,
// which must be called only at the signing boundary.
type Secret string

// String returns the redaction placeholder.
func (Secret) String() string { return "***" }

// MarshalJSON returns the redaction placeholder so secrets never serialize.
func (Secret) MarshalJSON() ([]byte, error) { return []byte(`"***"`), nil }

// GoString implements fmt.GoStringer so %#v (and spew-style dumps) cannot print
// the raw value, closing the one format verb String() does not cover.
func (Secret) GoString() string { return `config.Secret("***")` }

// Reveal returns the underlying value. Call only when signing.
func (s Secret) Reveal() string { return string(s) }

// SecretResolver turns a secret_source ("env:NAME") into a Secret.
type SecretResolver interface {
	Resolve(source string) (Secret, error)
}

// EnvResolver resolves "env:NAME" sources from the process environment.
type EnvResolver struct{}

// Resolve reads the env var named by an "env:NAME" source.
func (EnvResolver) Resolve(source string) (Secret, error) {
	const prefix = "env:"
	if !strings.HasPrefix(source, prefix) {
		return "", fmt.Errorf("config: secret source must be env:NAME, got %q", source)
	}
	name := strings.TrimPrefix(source, prefix)
	val, ok := os.LookupEnv(name)
	if !ok || val == "" {
		return "", fmt.Errorf("config: env var %q is empty or unset", name)
	}
	return Secret(val), nil
}
