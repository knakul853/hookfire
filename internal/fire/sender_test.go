package fire

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/knakul853/hookfire/internal/sign"
	"github.com/stretchr/testify/require"
)

func TestSendCapturesResponse(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		require.Equal(t, `{"hello":"world"}`, string(body))
		require.Equal(t, "sha256=abc", r.Header.Get("X-Hub-Signature-256"))
		w.WriteHeader(200)
		_, _ = w.Write([]byte(`{"ok":true}`))
	}))
	defer srv.Close()

	req, err := http.NewRequest(http.MethodPost, srv.URL, strings.NewReader(`{"hello":"world"}`))
	require.NoError(t, err)
	req.Header.Set("X-Hub-Signature-256", "sha256=abc")

	res, err := NewSender(Options{}).Send(context.Background(), req)
	require.NoError(t, err)
	require.Equal(t, 200, res.Status)
	require.Equal(t, `{"ok":true}`, string(res.Body))
	require.GreaterOrEqual(t, res.LatencyMS, int64(0))
	require.Equal(t, 11, res.Bytes)
}

func TestSendTransportError(t *testing.T) {
	req, _ := http.NewRequest(http.MethodPost, "http://127.0.0.1:1/nope", strings.NewReader("x"))
	_, err := NewSender(Options{}).Send(context.Background(), req)
	require.Error(t, err)
}

func TestEndToEndSignatureVerifies(t *testing.T) {
	const secret = "hookfire_test_secret"
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		mac := hmac.New(sha256.New, []byte(secret))
		mac.Write(body)
		want := "sha256=" + hex.EncodeToString(mac.Sum(nil))
		if r.Header.Get("X-Hub-Signature-256") != want {
			w.WriteHeader(401)
			return
		}
		w.WriteHeader(200)
		_, _ = w.Write([]byte(`{"ok":true}`))
	}))
	defer srv.Close()

	body := []byte(`{"hello":"world"}`)
	signer, err := sign.New(sign.SigningConfig{
		Scheme: "hmac", Algorithm: "sha256", Encoding: "hex",
		Basestring: "{{body}}", Output: "sha256={{sig}}", Header: "X-Hub-Signature-256",
	})
	require.NoError(t, err)
	sigHeaders, err := signer.Sign(body, sign.Options{Secret: secret, Timestamp: 1700000000})
	require.NoError(t, err)

	req, err := BuildRequest(RequestSpec{
		Method: "POST", URL: srv.URL, Body: body,
		Headers:          map[string]string{"Content-Type": "application/json"},
		SignatureHeaders: sigHeaders,
	})
	require.NoError(t, err)

	res, err := NewSender(Options{}).Send(context.Background(), req)
	require.NoError(t, err)
	require.Equal(t, 200, res.Status, "server must accept our signature")
}
