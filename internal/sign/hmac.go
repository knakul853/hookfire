package sign

import (
	"crypto/hmac"
	"crypto/sha1"
	"crypto/sha256"
	"crypto/sha512"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"hash"
	"net/http"
	"strconv"
	"strings"
)

type hmacSigner struct {
	cfg    SigningConfig
	newMAC func() hash.Hash
	encode func([]byte) string
}

func newHMAC(cfg SigningConfig) (Signer, error) {
	if cfg.Header == "" {
		return nil, fmt.Errorf("sign: hmac requires a header name")
	}
	newMAC, err := macFactory(cfg.Algorithm)
	if err != nil {
		return nil, err
	}
	enc, err := encoder(cfg.Encoding)
	if err != nil {
		return nil, err
	}
	return &hmacSigner{cfg: cfg, newMAC: newMAC, encode: enc}, nil
}

func macFactory(algo string) (func() hash.Hash, error) {
	switch algo {
	case "sha1":
		return sha1.New, nil
	case "sha256":
		return sha256.New, nil
	case "sha512":
		return sha512.New, nil
	default:
		return nil, fmt.Errorf("%w: algorithm %q", ErrUnsupportedScheme, algo)
	}
}

func encoder(enc string) (func([]byte) string, error) {
	switch enc {
	case "hex":
		return hex.EncodeToString, nil
	case "base64":
		return base64.StdEncoding.EncodeToString, nil
	default:
		return nil, fmt.Errorf("%w: encoding %q", ErrUnsupportedScheme, enc)
	}
}

func (s *hmacSigner) Sign(body []byte, opts SignOptions) (http.Header, error) {
	if opts.Secret == "" {
		return nil, ErrMissingSecret
	}
	ts := strconv.FormatInt(opts.Timestamp, 10)
	base := expand(s.cfg.Basestring, map[string]string{
		"body":      string(body),
		"timestamp": ts,
	})

	mac := hmac.New(s.newMAC, []byte(opts.Secret))
	mac.Write([]byte(base))
	sig := s.encode(mac.Sum(nil))

	value := expand(s.cfg.Output, map[string]string{"sig": sig, "timestamp": ts})

	h := http.Header{}
	h.Set(s.cfg.Header, value)
	for k, tmpl := range s.cfg.AuxHeaders {
		h.Set(k, expand(tmpl, map[string]string{"timestamp": ts}))
	}
	return h, nil
}

// expand replaces {{key}} tokens in tmpl with vars[key]. Unknown tokens are
// left as-is so a malformed template surfaces visibly rather than silently.
func expand(tmpl string, vars map[string]string) string {
	out := tmpl
	for k, v := range vars {
		out = strings.ReplaceAll(out, "{{"+k+"}}", v)
	}
	return out
}
