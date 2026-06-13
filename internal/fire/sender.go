package fire

import (
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"net/http"
	"time"
)

// Sender fires a prepared request and returns a captured Result.
type Sender interface {
	Send(ctx context.Context, req *http.Request) (Result, error)
}

// Options configures the http Sender.
type Options struct {
	Insecure bool          // skip TLS verification (CLI warns)
	Timeout  time.Duration // 0 means a sane default
}

type httpSender struct{ client *http.Client }

// NewSender builds the default http-backed Sender.
func NewSender(o Options) Sender {
	timeout := o.Timeout
	if timeout == 0 {
		timeout = 30 * time.Second
	}
	tr := &http.Transport{}
	if o.Insecure {
		tr.TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	}
	return &httpSender{client: &http.Client{Timeout: timeout, Transport: tr}}
}

func (s *httpSender) Send(ctx context.Context, req *http.Request) (Result, error) {
	start := time.Now()
	resp, err := s.client.Do(req.WithContext(ctx))
	if err != nil {
		return Result{}, fmt.Errorf("fire: send request: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return Result{}, fmt.Errorf("fire: read response: %w", err)
	}
	return Result{
		Status:    resp.StatusCode,
		LatencyMS: time.Since(start).Milliseconds(),
		Bytes:     len(body),
		Body:      body,
		Headers:   resp.Header,
	}, nil
}
