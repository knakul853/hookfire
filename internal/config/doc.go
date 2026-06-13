// Package config loads ~/.config/hookfire/config.toml, resolves target aliases
// to a URL + secret, and enforces secret hygiene. Secrets are referenced by
// env-var name, never stored in plaintext, and the Secret type redacts itself in
// every String/format/JSON path so it cannot leak into logs or output.
package config
