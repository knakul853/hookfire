# hookfire — Implementation Handoff Spec

> **Status:** Design approved, ready to implement. **Name:** `hookfire` is provisional — confirm before locking the Go module path, binary name, and config dir.
>
> **For the implementing session:** Start with `superpowers:writing-plans` to turn this into a task-by-task plan, then execute with `superpowers:subagent-driven-development` (or `executing-plans`). Every unit is built RED→GREEN with `superpowers:test-driven-development`. This document is the source of truth for *what* and *why*; the plan decides *order of steps*.

---

## 1. One-liner

A single static Go binary that fires **correctly-shaped, correctly-signed** synthetic webhook events at any URL. Think `stripe trigger`, but provider-agnostic — and providers are **data, not code**.

```
hookfire trigger github pull_request.opened -t local
```

---

## 2. Why this exists (motivation)

Testing an *inbound* webhook handler locally is annoying out of all proportion to its size. To exercise your `/webhook` you need an event that is **shaped exactly** like the provider's real payload **and signed** with the provider's exact scheme, delivered to localhost. Today people do one of three bad things:

1. **Trigger the real action** (push a commit, create a Stripe charge, open a Jira issue) and tunnel it via ngrok/smee — slow, side-effecty, needs real accounts/sandboxes, rate-limited, and you can't summon edge cases on demand.
2. **Hand-curl with a hand-computed HMAC** — and get the basestring/encoding/prefix subtly wrong (Slack's `v0:ts:body`, Stripe's `t=,v1=`), burn an hour, then **disable signature verification "just to test"** — which quietly ships as a security hole.
3. **Re-curl a saved payload** — whose signature is now stale, so again they bypass verification.

hookfire makes a real, signed event **one command**, so you test the handler *including its signature check* without the real action or hand-rolled crypto. **It removes the incentive to turn signature verification off.**

**Validation:** this tool has already been re-implemented three times, badly, in bash (the `neo-trigger`, Jira-`trigger`, and Vercel synthetic-event scripts at ProjectDiscovery), each hard-coding one provider's payload + signing. A real production bug there — Vercel sending `target=null` for previews — would have been reproducible as `hookfire trigger vercel deployment.ready --set payload.target=null -t local`, no deploy needed. It's dogfoodable on day one.

---

## 3. Goals / Non-goals

### v1 goals
- Fire a signed, realistically-shaped synthetic event at a target URL with one command.
- Providers (payload templates + signing scheme) are **declarative data** — adding one is a folder, not a Go change.
- Correct signing for the launch providers: **GitHub, Slack, Stripe, Vercel** (all HMAC variants), expressed purely as manifest data.
- Replay a saved JSON payload (re-signed).
- Tasteful, distinctive **HUD-pipeline** terminal output (render → sign → fire) with color on a TTY; clean plain/JSON output off-TTY or with `--json`.
- Production-grade engineering: godoc package docs, TDD, mockery mocks, focused files (<1000 lines, justify exceptions), DRY, deliberate log levels, built to extend.

### v1 non-goals (explicitly deferred — YAGNI)
- **Capture/tunnel** of real inbound events (that's ngrok/smee/Hookdeck's job; v2 at most).
- **Response assertions** / test-runner mode (`--expect`) — v2.
- **Asymmetric signing** (JWT/JWKS: Jira FIT, Vercel-OIDC) — the engine reserves a `jwt` scheme but v1 ships HMAC + none only.
- Interactive TUI, provider marketplace/registry, GUI.

---

## 4. Architecture & data flow

Single binary. Built-in providers are compiled in via `go:embed`; user/community providers are read from disk and merged over the built-ins by name.

```
trigger <provider> <event>
  └─ catalog.Lookup      → Manifest + EventTemplate (embedded ∪ filesystem)
  └─ render.Render       → interpolate {{vars}} + apply --set overrides → body bytes
  └─ sign.Sign           → compute signature headers from manifest.signing
  └─ fire.Send           → build HTTP request (method/headers/body) → POST target → Result
  └─ ui.Render           → HUD pipeline (TTY) | plain | JSON

replay <file.json>
  └─ (skip render; read file as body) → sign → fire → ui
```

Bottom-up dependency order (also the recommended build order): `sign` and `render` are pure and have no deps; `provider`, `config`, `fire` sit above; `cli` and `ui` are the top.

---

## 5. Provider contract (the heart of "data, not code")

A provider is a directory:

```
providers/github/
  manifest.yaml
  events/
    push.json
    pull_request.opened.json
    issue_comment.created.json
```

### 5.1 `manifest.yaml`

```yaml
name: github
description: GitHub webhooks
homepage: https://docs.github.com/webhooks

transport:
  method: POST
  content_type: application/json
  headers:                       # sent on every event
    User-Agent: "GitHub-Hookshot/hookfire"

signing:
  scheme: hmac                   # hmac | none   (jwt reserved for v2)
  algorithm: sha256              # sha1 | sha256 | sha512
  encoding: hex                  # hex | base64
  secret_source: "env:GITHUB_WEBHOOK_SECRET"   # default; overridable by flag/target
  basestring: "{{body}}"         # string that gets HMAC'd. vars: {{body}} {{timestamp}}
  output: "sha256={{sig}}"       # header value template. vars: {{sig}} {{timestamp}}
  header: "X-Hub-Signature-256"
  aux_headers: {}                # extra headers (templated with {{timestamp}})

events:                          # OPTIONAL per-event overrides; events also auto-discovered from events/*.json
  push:
    headers:
      X-GitHub-Event: push
  pull_request.opened:
    headers:
      X-GitHub-Event: pull_request   # note: action ("opened") lives in the body, not the header
```

- **Event discovery:** every `events/*.json` is an event named after the file (minus `.json`). The optional `events:` map adds/overrides per-event headers or metadata. Keep the common case zero-config.
- **Validation (fail closed):** unknown `scheme`/`algorithm`/`encoding`, missing `header` when `scheme != none`, or a `secret_source` that can't be parsed → manifest is invalid and the provider is rejected with a clear error.

### 5.2 Proof the four launch providers are pure data

| Provider | scheme | algo | encoding | basestring | output | header | aux |
|---|---|---|---|---|---|---|---|
| github | hmac | sha256 | hex | `{{body}}` | `sha256={{sig}}` | `X-Hub-Signature-256` | — |
| slack | hmac | sha256 | hex | `v0:{{timestamp}}:{{body}}` | `v0={{sig}}` | `X-Slack-Signature` | `X-Slack-Request-Timestamp: {{timestamp}}` |
| stripe | hmac | sha256 | hex | `{{timestamp}}.{{body}}` | `t={{timestamp}},v1={{sig}}` | `Stripe-Signature` | — |
| vercel | hmac | sha1 | hex | `{{body}}` | `{{sig}}` | `x-vercel-signature` | — |

If any future scheme can't be expressed by `{scheme, algorithm, encoding, basestring, output, header, aux_headers}`, **that is the engine failing** — extend the model deliberately, don't special-case in Go.

### 5.3 Provider resolution order (later overrides earlier, by `name`)
1. Embedded built-ins (`go:embed providers/`).
2. `$XDG_CONFIG_HOME/hookfire/providers/` (default `~/.config/hookfire/providers/`).
3. `./providers/` (current working directory).
4. `--providers-dir <path>` (highest).

A filesystem provider shadowing a built-in logs a `warn` (so users know they've overridden a shipped provider).

---

## 6. Signing engine (`internal/sign`)

The riskiest, most valuable unit — **build it first, test it hardest** (it gates the providers-as-data thesis).

```go
// Signer turns a rendered body into the headers a provider expects.
type Signer interface {
    // Sign returns the signature header(s) for body, or an error if the
    // secret is missing/invalid. It never mutates body and never logs the secret.
    Sign(body []byte, opts SignOptions) (http.Header, error)
}

type SignOptions struct {
    Secret    string    // resolved upstream; empty is an error unless scheme==none
    Timestamp int64     // unix seconds; injected (now) or overridden via --timestamp
}
```

**HMAC algorithm (the only concrete signer in v1 besides `none`):**
1. Render `basestring` with `{{body}}` (raw rendered bytes as string) and `{{timestamp}}`.
2. `mac = HMAC(algorithm, secret, basestring)`.
3. `sig = encode(mac)` where encode ∈ {hex lowercase, base64 std}.
4. Render `output` with `{{sig}}` and `{{timestamp}}` → header value.
5. Set `header: value` plus each rendered `aux_headers` entry.

`none` returns empty headers. `jwt` returns `ErrUnsupportedScheme` in v1 (reserved).

**Determinism for tests:** `Timestamp` is always injected from outside (never read the clock inside the signer), so golden vectors are reproducible.

---

## 7. CLI surface (`internal/cli`, cobra)

```
hookfire trigger <provider> <event> [flags]
hookfire replay  <file.json> --provider <p> [flags]
hookfire show    <provider> <event>            # = trigger --dry-run (render + sign, no fire)
hookfire list    [provider]                    # providers, or one provider's events
hookfire version
```

**Common flags:**
| Flag | Meaning |
|---|---|
| `-t, --target <alias>` | Resolve url + secret + (optional) provider from config. |
| `--url <url>` | Explicit target URL (overrides target's url). |
| `--secret-env <NAME>` | Read signing secret from this env var. |
| `--secret <value>` | Literal secret (discouraged → `warn`; prefer env). |
| `--set <path=value>` | Override a payload field by dot-path. Repeatable. |
| `--header <k=v>` | Add/override a request header. Repeatable. |
| `--timestamp <unix>` | Pin the signing timestamp (replay / expiry testing). |
| `--no-sign` | Send unsigned (for testing rejection paths). |
| `--dry-run` | Render + sign, print, **never** touch the network. |
| `--json` | Machine-readable output (implies no ANSI). |
| `--fail` | Exit non-zero on a non-2xx target response (curl `-f` style). |
| `--insecure` | Skip TLS verification (`warn`). |
| `-v, --verbose` | `-v` info, `-vv` debug — all to **stderr**. |
| `--providers-dir <path>` | Extra providers directory (highest precedence). |

**Exit codes:**
| Code | Meaning |
|---|---|
| 0 | Delivered (target response printed regardless of status, unless `--fail`). |
| 2 | Usage error: unknown provider/event (suggest near matches), bad flag. |
| 3 | Signing error: missing secret when signing required, unsupported scheme. |
| 4 | Transport error: could not reach target. |
| 22 | `--fail` set and target returned non-2xx. |
| 1 | Any other runtime error. |

---

## 8. Config & secrets (`internal/config`)

`~/.config/hookfire/config.toml` (respect `$XDG_CONFIG_HOME`):

```toml
[targets.local]
url        = "http://localhost:3000/webhook"
provider   = "github"                 # optional default provider for this alias
secret_env = "GITHUB_WEBHOOK_SECRET"

[targets.staging]
url        = "https://staging.example.com/hooks/github"
secret_env = "STAGING_GH_SECRET"
```

**Secret resolution precedence (highest first):** `--secret` → `--secret-env` → `target.secret_env` → `manifest.signing.secret_source`.

**Secret hygiene (hard rules):**
- Secrets are referenced **by env-var name**, never stored plaintext in config.
- The secret value MUST NOT appear in any output, log, or `String()`/`%v` of any struct. Add a `Secret` type whose `String()`/`MarshalJSON` returns `"***"`.
- `debug` may log the **basestring length** and **redacted** headers — never the secret or the raw signature-producing secret material.

---

## 9. Template rendering & `--set` (`internal/render`)

- Event JSON files contain `{{var}}` placeholders with realistic defaults baked in.
- Built-in vars: `{{timestamp}}` (unix s), `{{uuid}}`, `{{event}}`, `{{now}}` (RFC3339), `{{iso8601}}`.
- `--set <dot.path>=<value>` overrides a field in the parsed JSON before signing:
  - Dot-path navigates objects; `arr.0.field` indexes arrays.
  - Value type inference: `true/false`→bool, numeric→number, `null`→null, JSON literal (`{...}`/`[...]`)→parsed, else string. `--set-string` forces string.
  - Creating a missing intermediate path is allowed for objects; out-of-range array index is an error (exit 2).
- Render order: load template → interpolate `{{vars}}` → parse JSON → apply `--set` → marshal canonical bytes (these exact bytes are what gets signed and sent — sign the body you send, byte-for-byte).

---

## 10. Output / UI — the HUD pipeline (`internal/ui`)

The visual identity. Lean into tasteful color on a TTY (the user wants it); degrade cleanly everywhere else.

**Detection:** color + HUD only when stdout is a TTY **and** `--json` is off **and** `NO_COLOR` is unset. Otherwise plain or JSON.

**HUD layout (success):**
```
hookfire  github/pull_request.opened

 render  ▸ 1.2kb payload · 3 overrides
 sign    ▸ hmac-sha256 → X-Hub-Signature-256
 fire    ▸ POST localhost:3000/webhook
 ────────────────────────────────────────────
 ✓ 200 OK · 142ms · 18b   {"ok":true}
```

**Palette (lipgloss, adaptive light/dark):**
- header `hookfire` dim; `<provider>/<event>` bold + accent (cyan).
- step labels (`render/sign/fire`) dim grey; `▸` accent; values bright/cyan.
- separator: dim.
- result: `✓` green / `✗` red; status + latency + bytes bright; response snippet dim, truncated (~80 cols, full in `--verbose`).
- A subtle accent **gradient** on the `hookfire` header is a nice-to-have (lipgloss supports it) — keep it optional and cheap.

**States:**
- `--no-sign`: ` sign    ▸ skipped (--no-sign)` in dim/yellow.
- `--dry-run` / `show`: replace the fire line with ` fire    ▸ dry-run (not sent)` and print the rendered payload + final headers (signature included) below the separator.
- error (e.g. transport): ` ✗ connection refused · localhost:3000` in red; exit 4.
- spinner during `fire` only if it takes >150ms (avoid flicker on fast localhost).
- non-TTY plain mode: aligned `key: value` lines, no ANSI.
- `--json`: see schema below.

**`--json` schema (lock exact field names during the UI milestone):**
```json
{
  "provider": "github",
  "event": "pull_request.opened",
  "signed": true,
  "signature": { "header": "X-Hub-Signature-256", "scheme": "hmac-sha256" },
  "request":  { "method": "POST", "url": "http://localhost:3000/webhook", "headers": {"...": "..."}, "bytes": 1234 },
  "response": { "status": 200, "latency_ms": 142, "bytes": 18, "body": "{\"ok\":true}" },
  "dry_run": false
}
```
(Headers in JSON output redact the signing secret if it ever appears; the signature value itself is fine to include.)

---

## 11. Logging

This is a CLI: **stdout is the product** (HUD/plain/JSON), **stderr is for diagnostics**. Use `log/slog` to stderr, gated by verbosity. Never write logs to stdout (keeps `--json` pipeable).

| Level | Default? | Use for |
|---|---|---|
| error | yes (stderr) | unrecoverable: bad config, can't reach target, invalid manifest. Mirrored into HUD. |
| warn | yes | degraded but proceeding: `--secret` instead of env, filesystem provider shadowing a built-in, `--insecure`. |
| info | `-v` | lifecycle: provider resolved from dir X, target alias → url, body size. |
| debug | `-vv` | basestring length, **redacted** headers, template/`--set` resolution. Never secrets. |

---

## 12. Repo layout & package design

```
hookfire/
  cmd/hookfire/main.go              # wires cobra root; no logic
  internal/
    cli/        root.go trigger.go replay.go show.go list.go version.go
    provider/   catalog.go manifest.go event.go    # parse, validate, embed∪fs merge, discovery
    sign/       signer.go hmac.go none.go           # jwt.go reserved (v2)
    render/     render.go set.go
    fire/       sender.go result.go                 # http client + captured response
    config/     config.go target.go secret.go       # TOML, resolution, Secret type
    ui/         hud.go theme.go plain.go json.go tty.go
  providers/                                         # embedded built-ins
    github/  slack/  stripe/  vercel/
  mocks/                                             # mockery output
  testdata/                                          # golden signature vectors + golden UI output
  .mockery.yaml  .goreleaser.yaml  go.mod  README.md  LICENSE
```

**Standards (enforced in review):**
- Package-level godoc comment on every package; exported identifiers documented in godoc style.
- Files focused and <1000 lines; a larger file must carry a one-line justification.
- DRY: one signing engine, one render path, one HTTP sender — no per-provider code.
- No narration comments. Comments explain *why*, not *what the next line does*.
- Errors wrapped with context (`fmt.Errorf("...: %w", err)`); no swallowed errors.

---

## 13. Interfaces & mockery

Generate mocks with `mockery` (config in `.mockery.yaml`); never hand-roll.

| Interface | Package | Purpose / why mocked |
|---|---|---|
| `Catalog` | provider | `Lookup(provider, event) (Manifest, Event, error)`, `List()`. Mocked in cli tests. |
| `Signer` | sign | `Sign(body, opts) (http.Header, error)`. Mocked in cli/fire tests; real in sign tests. |
| `Sender` | fire | `Send(ctx, *http.Request) (Result, error)`. Mocked so cli tests never hit the network. |
| `SecretResolver` | config | `Resolve(source string) (Secret, error)`. Mocked to test precedence without real env. |

Run `mockery` after adding/changing any interface (don't drift).

---

## 14. Testing strategy (RED→GREEN throughout)

- **sign (golden vectors):** for each launch provider, a fixed `{body, secret, timestamp}` → the exact expected header value, cross-checked against the provider's real verification semantics. Write the failing test first. This is the conformance core.
- **render:** `{{var}}` interpolation; `--set` dot-path (nested object create, array index, type inference, `--set-string`); bad-path error.
- **provider:** manifest parse + validation (reject unknown scheme / missing header); event auto-discovery; fs-overrides-embedded precedence + shadow warning.
- **config:** TOML parse; target resolution; secret precedence; `Secret.String()`/JSON redaction.
- **fire:** request shape via mock `Sender`; **integration** via `httptest.Server` that *verifies the received signature* — the same check exposed later as a `verify` conformance command for community providers.
- **ui:** golden-output tests (ANSI stripped) for HUD / plain / JSON / dry-run / error states; TTY vs non-TTY branch.
- **e2e:** build the binary; `trigger github push --dry-run --json` → assert JSON shape; unknown provider → exit 2.

Never weaken a real check to make a test pass. Tests assert behavior, not implementation.

---

## 15. Distribution

- **goreleaser:** cross-compile darwin/linux/windows × amd64/arm64 → GitHub Releases; Homebrew tap; `go install ./cmd/hookfire`; `curl … | sh` install script.
- **README:** hero one-liners, an asciinema of the HUD, and a "**add a provider = add a folder**" contributing guide (the low-friction first PR that grows the project).
- **License:** Apache-2.0.
- **Shell completions:** cobra-generated (`completion` command) — cheap, add at the end.

---

## 16. Implementation roadmap (milestones → feed to writing-plans)

Bottom-up, pure→impure, each milestone independently testable.

- **M0 — Scaffold:** go.mod, cobra root + `version`, lint+test CI, `.goreleaser.yaml` skeleton, `.mockery.yaml`. Green CI on an empty binary.
- **M1 — Signing engine + golden vectors** (riskiest first; proves the thesis). `sign.Signer`, `hmac`, `none`; golden tests for github/slack/stripe/vercel.
- **M2 — Provider catalog + the 4 built-in provider folders.** Manifest parse/validate, `go:embed`, fs merge + discovery.
- **M3 — Render + `--set`.**
- **M4 — Config + secret resolution** (incl. `Secret` redaction type).
- **M5 — Fire** (http `Sender` + `Result`, integration test with signature-verifying server).
- **M6 — UI** (HUD + plain + JSON + theme + TTY detect; golden output tests). Lock the `--json` schema here.
- **M7 — CLI wiring:** `trigger`, `show`, `list`, `replay`; flags; exit codes.
- **M8 — Polish & ship:** e2e, README + asciinema, goreleaser/brew/install.sh, completions, the `verify` conformance command.

---

## 17. Open questions for the implementer / owner

1. **Final name** (`hookfire` provisional) — fixes module path, binary, config dir, tap name. Decide before M0.
2. **Homebrew tap location** — personal tap vs org.
3. **`--json` field names** — proposed in §10; lock during M6.

## 18. Deferred to v2+ (do not build in v1)

Capture/tunnel of real events · response assertions (`--expect`) · asymmetric `jwt` signing (Jira FIT, Vercel-OIDC) · `providers add <git-url>` · optional `hookfire tui` · provider registry/index.
