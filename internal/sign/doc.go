// Package sign turns a rendered request body into the signature header(s) a
// provider expects, driven entirely by a provider's declarative SigningConfig.
//
// It is the conformance core of hookfire: one HMAC engine (plus a no-op "none"
// signer) expresses every launch provider's scheme as data. The signer is pure
// and deterministic — the timestamp is injected via SignOptions and never read
// from the clock — so golden vectors are reproducible. It never mutates the
// body and never logs the secret.
package sign
