package ui

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/require"
)

func sampleView() View {
	return View{
		Provider: "github", Event: "pull_request.opened", Signed: true,
		Signature: SignatureView{Header: "X-Hub-Signature-256", Scheme: "hmac-sha256"},
		Request: RequestView{Method: "POST", URL: "http://localhost:3000/webhook",
			Headers: map[string]string{"X-Hub-Signature-256": "sha256=abc"}, Bytes: 1234},
		Response: &ResponseView{Status: 200, LatencyMS: 142, Bytes: 18, Body: `{"ok":true}`},
		DryRun:   false,
	}
}

func TestJSONSchemaLocked(t *testing.T) {
	var buf bytes.Buffer
	require.NoError(t, RenderJSON(&buf, sampleView()))
	require.JSONEq(t, `{
	  "provider":"github",
	  "event":"pull_request.opened",
	  "signed":true,
	  "signature":{"header":"X-Hub-Signature-256","scheme":"hmac-sha256"},
	  "request":{"method":"POST","url":"http://localhost:3000/webhook","headers":{"X-Hub-Signature-256":"sha256=abc"},"bytes":1234},
	  "response":{"status":200,"latency_ms":142,"bytes":18,"body":"{\"ok\":true}"},
	  "dry_run":false
	}`, buf.String())
}

func TestJSONDryRunOmitsResponse(t *testing.T) {
	v := sampleView()
	v.DryRun = true
	v.Response = nil
	var buf bytes.Buffer
	require.NoError(t, RenderJSON(&buf, v))
	require.Contains(t, buf.String(), `"dry_run": true`)
	require.Contains(t, buf.String(), `"response": null`)
}
