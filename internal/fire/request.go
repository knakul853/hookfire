package fire

import (
	"bytes"
	"fmt"
	"net/http"
)

// RequestSpec is everything needed to build the outbound request. SignatureHeaders
// are applied last so they cannot be clobbered by transport/--header values; each
// signature header replaces any same-named user header while keeping all of its
// own values (hence []string).
type RequestSpec struct {
	Method           string
	URL              string
	Body             []byte
	Headers          map[string]string
	SignatureHeaders map[string][]string
}

// BuildRequest assembles the HTTP request: transport+user headers first, then
// signature headers, with the body length pinned to the exact signed bytes.
func BuildRequest(spec RequestSpec) (*http.Request, error) {
	method := spec.Method
	if method == "" {
		method = http.MethodPost
	}
	req, err := http.NewRequest(method, spec.URL, bytes.NewReader(spec.Body))
	if err != nil {
		return nil, fmt.Errorf("fire: build request: %w", err)
	}
	req.ContentLength = int64(len(spec.Body))
	for k, v := range spec.Headers {
		req.Header.Set(k, v)
	}
	for k, vals := range spec.SignatureHeaders {
		// Assign the whole slice so a multi-value signature header keeps every
		// value, and so the signature replaces any same-named user header.
		req.Header[http.CanonicalHeaderKey(k)] = vals
	}
	return req, nil
}
