package fire

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/knakul853/hookfire/internal/sign"
	"github.com/stretchr/testify/require"
)

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
