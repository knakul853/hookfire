# hookfire

[![ci](https://github.com/knakul853/hookfire/actions/workflows/ci.yml/badge.svg)](https://github.com/knakul853/hookfire/actions/workflows/ci.yml)
[![Go Reference](https://pkg.go.dev/badge/github.com/knakul853/hookfire.svg)](https://pkg.go.dev/github.com/knakul853/hookfire)
[![Go Report Card](https://goreportcard.com/badge/github.com/knakul853/hookfire)](https://goreportcard.com/report/github.com/knakul853/hookfire)
[![License](https://img.shields.io/badge/license-Apache--2.0-blue.svg)](LICENSE)

A single static Go binary that fires **correctly-shaped, correctly-signed** synthetic webhook events at any URL. Think `stripe trigger`, but provider-agnostic — and providers are **data, not code**.

```sh
hookfire trigger github pull_request.opened -t local
```

Testing an inbound webhook handler locally is annoying out of proportion to its size: you need a payload shaped exactly like the provider's and signed with the provider's exact scheme. The usual workarounds (trigger the real action over a tunnel, or hand-roll an HMAC and then disable signature verification "just to test") are slow, side-effecty, or quietly ship a security hole. hookfire makes a real, signed event one command — so you test the handler **including its signature check**.

## Install

```sh
# from source (requires Go 1.24+)
go install github.com/knakul853/hookfire/cmd/hookfire@latest

# install script — downloads the latest release binary (Linux/macOS)
curl -fsSL https://raw.githubusercontent.com/knakul853/hookfire/main/install.sh | sh
```

Or grab a prebuilt binary for your platform from the [Releases](https://github.com/knakul853/hookfire/releases) page. A Homebrew tap is planned.

## Quickstart

```sh
# Fire a signed GitHub push at localhost (secret from $GITHUB_WEBHOOK_SECRET):
export GITHUB_WEBHOOK_SECRET=whsec_test
hookfire trigger github push --url http://localhost:3000/webhook

# Render + sign without sending, and inspect the payload + headers:
hookfire show stripe payment_intent.succeeded --secret whsec_test

# Override payload fields by dot-path (reproduce that one broken edge case):
hookfire trigger vercel deployment.ready --set payload.target=null --url http://localhost:3000/hook --secret s

# Machine-readable output for scripts/CI (stdout stays clean; logs go to stderr):
hookfire trigger github push --dry-run --json --url http://x --secret s

# Re-sign and replay a saved payload verbatim:
hookfire replay ./captured.json --provider github --url http://localhost:3000/webhook --secret s

# List what's available:
hookfire list            # providers
hookfire list github     # one provider's events
```

The HUD pipeline on a TTY:

```
hookfire  github/pull_request.opened

 render ▸ 1.2kb payload · 3 overrides
 sign   ▸ hmac-sha256 → X-Hub-Signature-256
 fire   ▸ POST localhost:3000/webhook
 ────────────────────────────────────────────
 ✓ 200 OK · 142ms · 18b   {"ok":true}
```

Off a TTY (or with `--json` / `NO_COLOR`) it degrades to clean plain text or the locked JSON schema.

## Launch providers

`github`, `slack`, `stripe`, `vercel` — every HMAC variant, expressed purely as manifest data. Each provider's signing is `{scheme, algorithm, encoding, basestring, output, header, aux_headers}`; there is **no per-provider Go code**.

## Secrets

Secrets are referenced by **environment-variable name**, never stored in config. Resolution precedence (highest first): `--secret` → `--secret-env` → a target's `secret_env` → the manifest's `secret_source`. Prefer `--secret-env` over `--secret` (the latter leaks into shell history, and hookfire warns). The secret value never appears in any output, log, or JSON — it is redacted everywhere.

## Config

`~/.config/hookfire/config.toml` (respects `$XDG_CONFIG_HOME`) defines target aliases:

```toml
[targets.local]
url        = "http://localhost:3000/webhook"
provider   = "github"
secret_env = "GITHUB_WEBHOOK_SECRET"
```

Then: `hookfire trigger github push -t local`.

## Add a provider = add a folder

Providers are data. To add one, drop a folder under `~/.config/hookfire/providers/` (or `./providers/`, or pass `--providers-dir`):

```
myprovider/
  manifest.yaml
  events/
    something.happened.json
```

`manifest.yaml`:

```yaml
name: myprovider
transport:
  method: POST
  content_type: application/json
signing:
  scheme: hmac          # hmac | none
  algorithm: sha256     # sha1 | sha256 | sha512
  encoding: hex         # hex | base64
  secret_source: "env:MYPROVIDER_SECRET"
  basestring: "{{body}}"          # vars: {{body}} {{timestamp}}
  output: "sha256={{sig}}"        # vars: {{sig}} {{timestamp}}
  header: "X-My-Signature"
```

Each `events/*.json` is an event named after the file. Payloads may use `{{timestamp}}`, `{{uuid}}`, `{{event}}`, `{{now}}`, `{{iso8601}}`. Filesystem providers override built-ins by name (with a warning). Confirm your provider signs correctly:

```sh
hookfire verify ./providers
```

If a real provider's scheme can't be expressed by the manifest fields, that's the engine failing — open an issue rather than special-casing it.

## Exit codes

| Code | Meaning |
|---|---|
| 0 | Delivered (response printed regardless of status, unless `--fail`). |
| 2 | Usage error: unknown provider/event (suggests near matches), bad flag. |
| 3 | Signing error: missing secret when signing required, unsupported scheme. |
| 4 | Transport error: could not reach the target. |
| 22 | `--fail` set and the target returned a non-2xx response. |
| 1 | Any other runtime error. |

## Development

```sh
go test ./... -race          # unit + integration
go test -tags e2e ./e2e/...  # end-to-end binary tests
golangci-lint run ./...
```

Mocks are generated (never hand-rolled): `go run github.com/vektra/mockery/v2@v2.53.4`.

## License

Apache-2.0.
