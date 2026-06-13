package provider

import (
	"testing"

	"github.com/stretchr/testify/require"
)

const githubManifest = `
name: github
description: GitHub webhooks
transport:
  method: POST
  content_type: application/json
  headers:
    User-Agent: "GitHub-Hookshot/hookfire"
signing:
  scheme: hmac
  algorithm: sha256
  encoding: hex
  secret_source: "env:GITHUB_WEBHOOK_SECRET"
  basestring: "{{body}}"
  output: "sha256={{sig}}"
  header: "X-Hub-Signature-256"
events:
  pull_request.opened:
    headers:
      X-GitHub-Event: pull_request
`

func TestParseManifest(t *testing.T) {
	m, err := ParseManifest([]byte(githubManifest))
	require.NoError(t, err)
	require.Equal(t, "github", m.Name)
	require.Equal(t, "POST", m.Transport.Method)
	require.Equal(t, "application/json", m.Transport.ContentType)
	require.Equal(t, "GitHub-Hookshot/hookfire", m.Transport.Headers["User-Agent"])
	require.Equal(t, "hmac", m.Signing.Scheme)
	require.Equal(t, "env:GITHUB_WEBHOOK_SECRET", m.Signing.SecretSource)
	require.Equal(t, "pull_request", m.Events["pull_request.opened"].Headers["X-GitHub-Event"])
}

func TestValidateRejectsUnknownScheme(t *testing.T) {
	m, _ := ParseManifest([]byte("name: x\nsigning:\n  scheme: magic\n"))
	require.Error(t, m.Validate())
}

func TestValidateRejectsMissingHeaderWhenSigning(t *testing.T) {
	m, _ := ParseManifest([]byte("name: x\nsigning:\n  scheme: hmac\n  algorithm: sha256\n  encoding: hex\n"))
	require.Error(t, m.Validate())
}

func TestValidateRejectsBadSecretSource(t *testing.T) {
	m, _ := ParseManifest([]byte("name: x\nsigning:\n  scheme: hmac\n  algorithm: sha256\n  encoding: hex\n  header: X\n  secret_source: \"vault:foo\"\n"))
	require.Error(t, m.Validate())
}

func TestValidateRejectsMissingName(t *testing.T) {
	m, _ := ParseManifest([]byte("signing:\n  scheme: none\n"))
	require.Error(t, m.Validate())
}

func TestValidateAcceptsGithub(t *testing.T) {
	m, _ := ParseManifest([]byte(githubManifest))
	require.NoError(t, m.Validate())
}

func TestValidateAcceptsNoneWithoutHeader(t *testing.T) {
	m, _ := ParseManifest([]byte("name: x\nsigning:\n  scheme: none\n"))
	require.NoError(t, m.Validate())
}
