// Package provider loads webhook provider definitions — a manifest.yaml plus an
// events/ folder of JSON payload templates — and exposes them through a Catalog.
//
// Providers are data, not code: built-ins are embedded via go:embed and
// filesystem providers merge over them by name, so adding a provider is adding
// a folder. Manifests are validated fail-closed: an unknown scheme/algorithm/
// encoding, a missing header when the scheme needs one, or an unparseable
// secret_source rejects the provider with a clear error.
package provider
