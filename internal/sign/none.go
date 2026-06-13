package sign

import "net/http"

// noneSigner implements the "none" scheme: it signs nothing.
type noneSigner struct{}

func (noneSigner) Sign(_ []byte, _ SignOptions) (http.Header, error) {
	return http.Header{}, nil
}
