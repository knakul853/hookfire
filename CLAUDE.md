# CLAUDE.md — hookfire

A Go CLI that fires correctly-signed synthetic webhook events at a URL. **Thesis: providers are data, not code.** A provider is a folder (`manifest.yaml` + `events/*.json`); the signing engine is generic over the manifest. Never add per-provider Go branching (`if provider == "github"`) — if a real scheme can't be expressed by the manifest fields, extend the *engine*, not special-case it.

## Layout
- `internal/sign` — HMAC engine + `none`; pure, deterministic (timestamp injected, never reads the clock).
- `internal/render` — `{{var}}` interpolation + `--set` dot-path → canonical JSON bytes (these exact bytes are signed & sent).
- `internal/provider` — manifest parse/validate, `go:embed builtin/`, fs-merge catalog.
- `internal/config` — `Secret` (redacts), `EnvResolver`, secret-precedence resolver, TOML targets.
- `internal/fire` — `Sender` + request builder.
- `internal/ui` — HUD / plain / locked `--json`; mode by TTY + `NO_COLOR` + `--json`.
- `internal/cli` — cobra commands + the shared `runPipeline`/`fireResult` path.
- `providers/builtin/<name>/` — the 4 shipped providers (data only).

## Non-negotiables (these have bitten us)
- **TDD RED→GREEN.** Write the failing test first. **Never edit a golden signature value to make a test pass** — golden vectors are independent oracles (cross-checked vs GitHub's published vector); if one fails, the *code* is wrong.
- **Secret hygiene.** `config.Secret` redacts in `String`/`%v`/`%s`/`%#v`/JSON. Only `.Reveal()` exposes it, and only at the signing call. Never log/print/format the revealed value; slog the *source name* only.
- **stdout = product, stderr = diagnostics.** Command output (HUD/plain/JSON) goes to `cmd.OutOrStdout()`. slog → stderr only (so `--json` stays pipeable). **Do NOT use cobra `cmd.Print*` for product output — it routes to stderr.**
- **Mocks are generated, never hand-written:** `go run github.com/vektra/mockery/v2@v2.53.4` (PATH may shadow it with an incompatible v3). Regenerate after changing any interface.
- **Go 1.24** is pinned in `go.mod`. Don't let `go get` bump the `go`/`toolchain` directive; pin deps to 1.24-compatible versions.
- **Lint is golangci-lint v2** (`.golangci.yml` is v2 schema); CI uses `golangci-lint-action@v7`. No speculative `//nolint`.
- Tests that hit config resolution must `t.Setenv("XDG_CONFIG_HOME", t.TempDir())` to stay hermetic (don't read the dev's real `~/.config`).
- **One test file per source file:** `foo.go` → `foo_test.go`, holding the tests for the functions defined in `foo.go`. No per-method files (`method_test.go`) and no grab-bag files (`extra_test.go`, `coverage_test.go`) — they drift and rot. Shared test helpers live once in the test file of the source they most belong to.

## Commands
`make test` (race) · `make e2e` · `make lint` · `make cover` · `make verify` · `make mocks` · `make build`.

## Workflow
`main` is protected (PR-only, CI must pass). Branch → commit → push → `gh pr create --fill` → self-merge once `test` + `lint` are green. Conventional commit messages (`feat(sign):`, `fix(cli):`, …).

## Adding a provider
Use the **add-provider** skill (`.claude/skills/add-provider/`). Short version: drop a folder under `providers/builtin/<name>/`, add `manifest.yaml` + `events/*.json`, add a golden vector to `internal/provider/conformance_test.go`, run `make verify`. No Go changes.
