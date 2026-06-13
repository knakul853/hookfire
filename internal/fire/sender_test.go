package fire

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

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
