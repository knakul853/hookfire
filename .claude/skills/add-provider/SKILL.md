---
name: add-provider
description: Use when adding or modifying a webhook provider in hookfire (a new provider like GitHub/Slack/Stripe, or editing a built-in's manifest/events). Covers the manifest schema, signing fields, event templates, the golden-vector conformance test, and verification — all data, no Go code.
---

# Adding a provider to hookfire

A provider is **data, not code**. You add a folder; you do **not** write or change any Go in `internal/`. If a provider's real signing scheme can't be expressed by the manifest fields below, that is the *engine* failing — stop and extend `internal/sign`, don't special-case in Go.

## 1. Create the folder

Built-ins live in `internal/provider/builtin/<name>/` (embedded via `go:embed`). User/community providers can instead go in `~/.config/hookfire/providers/`, `./providers/`, or `--providers-dir` (later overrides earlier by name).

```
<name>/
  manifest.yaml
  events/
    some.event.json
```

## 2. Write `manifest.yaml`

```yaml
name: <name>
description: <human label>
homepage: <docs url>
transport:
  method: POST
  content_type: application/json
  headers:                       # sent on every event (optional)
    User-Agent: "..."
signing:
  scheme: hmac                   # hmac | none   (jwt is reserved, v2)
  algorithm: sha256              # sha1 | sha256 | sha512
  encoding: hex                  # hex | base64
  secret_source: "env:<NAME>"    # default secret env var; optional (flags/targets can supply it)
  basestring: "{{body}}"         # string that is HMAC'd. vars: {{body}} {{timestamp}}
  output: "sha256={{sig}}"       # header VALUE template. vars: {{sig}} {{timestamp}}
  header: "X-...-Signature"      # header NAME for output (required unless scheme: none)
  aux_headers: {}                # extra headers, templated with {{timestamp}}
events:                          # OPTIONAL per-event header overrides
  some.event:
    headers: { X-Event: some }
```

The four shipped schemes, all pure data — match this shape:

| provider | algo | encoding | basestring | output | header | aux |
|---|---|---|---|---|---|---|
| github | sha256 | hex | `{{body}}` | `sha256={{sig}}` | `X-Hub-Signature-256` | — |
| slack | sha256 | hex | `v0:{{timestamp}}:{{body}}` | `v0={{sig}}` | `X-Slack-Signature` | `X-Slack-Request-Timestamp: {{timestamp}}` |
| stripe | sha256 | hex | `{{timestamp}}.{{body}}` | `t={{timestamp}},v1={{sig}}` | `Stripe-Signature` | — |
| vercel | sha1 | hex | `{{body}}` | `{{sig}}` | `x-vercel-signature` | — |

Validation is fail-closed: unknown scheme/algorithm/encoding, missing `header` when signing, or a non-`env:` `secret_source` rejects the provider.

## 3. Write event templates

Each `events/*.json` is an event named after the file (minus `.json`). Shape them like the provider's real payload. Available built-in vars: `{{timestamp}}` (unix s), `{{uuid}}`, `{{event}}`, `{{now}}` (RFC3339), `{{iso8601}}`. The file is raw bytes until rendered, so it only needs to be valid JSON *after* token substitution.

## 4. Add a golden conformance test (built-ins only)

For an embedded provider, add an entry to `internal/provider/conformance_test.go` proving the manifest reproduces an independently-computed signature for the fixed fixture (`secret=hookfire_test_secret`, `body={"hello":"world"}`, `timestamp=1700000000`). Compute the expected value with an **independent** tool (Python/openssl) — never copy it from hookfire's own output:

```python
import hmac, hashlib  # github example
hmac.new(b"hookfire_test_secret", b'{"hello":"world"}', hashlib.sha256).hexdigest()
```

## 5. Verify

```sh
make verify                 # signs the golden body for every provider; expects ✓
go test ./internal/provider/ -race
hookfire verify ./providers # for a filesystem provider dir
```

Then live-fire it: `hookfire trigger <name> <event> --url <local-receiver> --secret <s> --fail` (exit 0 = an independent verifier accepted the signature; exit 22 = mismatch).

## Rules
- No Go changes in `internal/` for a new provider.
- Don't edit a golden value to make a test pass — fix the manifest.
- `make lint` and `make test` must stay green; open a PR (main is protected).
