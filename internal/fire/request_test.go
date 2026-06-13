package fire

import (
	"io"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestBuildRequest(t *testing.T) {
	req, err := BuildRequest(RequestSpec{
		Method:           "POST",
		URL:              "http://localhost:3000/webhook",
		Body:             []byte(`{"a":1}`),
		Headers:          map[string]string{"Content-Type": "application/json", "X-GitHub-Event": "push"},
		SignatureHeaders: map[string][]string{"X-Hub-Signature-256": {"sha256=abc"}},
	})
	require.NoError(t, err)
	require.Equal(t, "POST", req.Method)
	require.Equal(t, "application/json", req.Header.Get("Content-Type"))
	require.Equal(t, "push", req.Header.Get("X-GitHub-Event"))
	require.Equal(t, "sha256=abc", req.Header.Get("X-Hub-Signature-256"))
	body, _ := io.ReadAll(req.Body)
	require.Equal(t, `{"a":1}`, string(body))
	require.Equal(t, int64(7), req.ContentLength)
}

func TestBuildRequestSignatureHeadersMultiValueAndReplace(t *testing.T) {
	req, err := BuildRequest(RequestSpec{
		URL:  "http://localhost/webhook",
		Body: []byte(`{}`),
		// A user header that the signature must override, not merely append to.
		Headers:          map[string]string{"X-Sig": "user-set"},
		SignatureHeaders: map[string][]string{"X-Sig": {"a", "b"}},
	})
	require.NoError(t, err)
	require.Equal(t, []string{"a", "b"}, req.Header.Values("X-Sig"))
}
