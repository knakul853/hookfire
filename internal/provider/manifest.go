package provider

import (
	"fmt"

	"github.com/knakul853/hookfire/internal/sign"
	"gopkg.in/yaml.v3"
)

// Manifest is the parsed manifest.yaml of a provider.
type Manifest struct {
	Name        string                   `yaml:"name"`
	Description string                   `yaml:"description"`
	Homepage    string                   `yaml:"homepage"`
	Transport   Transport                `yaml:"transport"`
	Signing     Signing                  `yaml:"signing"`
	Events      map[string]EventManifest `yaml:"events"`
}

// Transport describes how every event for this provider is delivered.
type Transport struct {
	Method      string            `yaml:"method"`
	ContentType string            `yaml:"content_type"`
	Headers     map[string]string `yaml:"headers"`
}

// Signing is the manifest signing block; SigningConfig converts it for the engine.
type Signing struct {
	Scheme       string            `yaml:"scheme"`
	Algorithm    string            `yaml:"algorithm"`
	Encoding     string            `yaml:"encoding"`
	SecretSource string            `yaml:"secret_source"`
	Basestring   string            `yaml:"basestring"`
	Output       string            `yaml:"output"`
	Header       string            `yaml:"header"`
	AuxHeaders   map[string]string `yaml:"aux_headers"`
}

// EventManifest holds optional per-event overrides from the manifest events: map.
type EventManifest struct {
	Headers map[string]string `yaml:"headers"`
}

// ParseManifest decodes manifest.yaml bytes. It does not validate; call Validate.
func ParseManifest(data []byte) (Manifest, error) {
	var m Manifest
	if err := yaml.Unmarshal(data, &m); err != nil {
		return Manifest{}, fmt.Errorf("provider: parse manifest: %w", err)
	}
	return m, nil
}

// SigningConfig converts the manifest signing block into the engine's config.
func (m Manifest) SigningConfig() sign.SigningConfig {
	return sign.SigningConfig{
		Scheme:     m.Signing.Scheme,
		Algorithm:  m.Signing.Algorithm,
		Encoding:   m.Signing.Encoding,
		Basestring: m.Signing.Basestring,
		Output:     m.Signing.Output,
		Header:     m.Signing.Header,
		AuxHeaders: m.Signing.AuxHeaders,
	}
}

// Validate enforces fail-closed manifest rules. It uses sign.New to confirm the
// signing block is actually constructable, so the catalog never serves a
// provider whose signer would error at fire time.
func (m Manifest) Validate() error {
	if m.Name == "" {
		return fmt.Errorf("provider: manifest missing name")
	}
	// An absent secret_source is allowed: the secret may instead come from
	// --secret, --secret-env, or a target alias at fire time (SPEC §8). When
	// present, it must be a parseable env:NAME reference.
	if m.Signing.SecretSource != "" {
		if _, err := parseSecretSource(m.Signing.SecretSource); err != nil {
			return fmt.Errorf("provider %q: %w", m.Name, err)
		}
	}
	if _, err := sign.New(m.SigningConfig()); err != nil {
		return fmt.Errorf("provider %q: invalid signing: %w", m.Name, err)
	}
	return nil
}

// parseSecretSource accepts only "env:NAME" in v1 (the engine reserves others).
func parseSecretSource(src string) (envVar string, err error) {
	const prefix = "env:"
	if len(src) <= len(prefix) || src[:len(prefix)] != prefix {
		return "", fmt.Errorf("provider: secret_source must be env:NAME, got %q", src)
	}
	return src[len(prefix):], nil
}
