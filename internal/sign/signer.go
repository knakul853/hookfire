package sign

import (
	"errors"
	"net/http"
)

// ErrUnsupportedScheme is returned for any scheme other than "hmac" or "none".
// "jwt" is reserved for v2 and returns this error in v1.
var ErrUnsupportedScheme = errors.New("sign: unsupported scheme")

// ErrMissingSecret is returned when a signing scheme requires a secret but
// Options.Secret is empty.
var ErrMissingSecret = errors.New("sign: missing secret")

// ErrInvalidConfig is returned when a scheme is selected but its config is
// incomplete (e.g. hmac without a header name), as distinct from a scheme that
// is not supported at all (ErrUnsupportedScheme).
var ErrInvalidConfig = errors.New("sign: invalid signing config")

// SigningConfig is the declarative signing description from a provider manifest.
// All fields except Scheme are ignored when Scheme == "none".
type SigningConfig struct {
	Scheme     string            // hmac | none (jwt reserved)
	Algorithm  string            // sha1 | sha256 | sha512
	Encoding   string            // hex | base64
	Basestring string            // template over {{body}} {{timestamp}}
	Output     string            // header value template over {{sig}} {{timestamp}}
	Header     string            // header name for Output
	AuxHeaders map[string]string // extra headers, templated with {{timestamp}}
}

// Options carries the per-call inputs that are resolved upstream.
type Options struct {
	Secret    string // resolved secret; empty is an error unless Scheme=="none"
	Timestamp int64  // unix seconds; injected (now) or pinned via --timestamp
}

// Signer turns a rendered body into the headers a provider expects.
type Signer interface {
	// Sign returns the signature header(s) for body, or an error if the secret
	// is missing/invalid. It never mutates body and never logs the secret.
	Sign(body []byte, opts Options) (http.Header, error)
}

// New builds the Signer for cfg.Scheme, validating algorithm/encoding for hmac.
func New(cfg SigningConfig) (Signer, error) {
	switch cfg.Scheme {
	case "none":
		return noneSigner{}, nil
	case "hmac":
		return newHMAC(cfg)
	default:
		return nil, ErrUnsupportedScheme
	}
}
