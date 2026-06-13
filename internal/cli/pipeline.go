package cli

import (
	"context"
	"net/http"
	"time"

	"github.com/knakul853/hookfire/internal/config"
	"github.com/knakul853/hookfire/internal/fire"
	"github.com/knakul853/hookfire/internal/provider"
	"github.com/knakul853/hookfire/internal/render"
	"github.com/knakul853/hookfire/internal/sign"
	"github.com/knakul853/hookfire/internal/ui"
)

// pipelineDeps are the injectable collaborators; tests supply mocks and may
// leave Now/UUID nil to get the real implementations.
type pipelineDeps struct {
	Catalog provider.Catalog
	Sender  fire.Sender
	Now     func() int64
	UUID    func() string
}

// pipelineInput is one fully-resolved request to render, sign, and fire.
type pipelineInput struct {
	Provider     string
	Event        string
	URL          string
	Secret       config.Secret
	Timestamp    int64
	Sets         []render.Set
	ExtraHeaders map[string]string
	NoSign       bool
	DryRun       bool
	Fail         bool
}

// runPipeline resolves the event, renders the body, signs it (unless NoSign or
// scheme==none), and (unless DryRun) fires it. The view is populated even on a
// transport error or --fail non-2xx so the caller can still render it.
func runPipeline(deps pipelineDeps, in pipelineInput) (ui.View, error) {
	now := deps.Now
	if now == nil {
		now = realNow
	}
	genUUID := deps.UUID
	if genUUID == nil {
		genUUID = newUUID
	}

	manifest, event, err := deps.Catalog.Lookup(in.Provider, in.Event)
	if err != nil {
		return ui.View{}, err
	}

	ts := in.Timestamp
	if ts == 0 {
		ts = now()
	}
	vars := render.Vars{
		Timestamp: ts,
		UUID:      genUUID(),
		Event:     in.Event,
		Now:       time.Unix(ts, 0).UTC().Format(time.RFC3339),
	}
	body, err := render.Render(event.Template, vars, in.Sets)
	if err != nil {
		return ui.View{}, err
	}

	method := manifest.Transport.Method
	if method == "" {
		method = http.MethodPost
	}
	headers := event.Headers()
	for k, v := range in.ExtraHeaders {
		headers[k] = v
	}

	signed := !in.NoSign && manifest.Signing.Scheme != "none"
	sigHeaders := http.Header{}
	var sigView ui.SignatureView
	if signed {
		signer, serr := sign.New(manifest.SigningConfig())
		if serr != nil {
			return ui.View{}, serr
		}
		if in.Secret.Reveal() == "" {
			return ui.View{}, sign.ErrMissingSecret
		}
		sigHeaders, serr = signer.Sign(body, sign.Options{Secret: in.Secret.Reveal(), Timestamp: ts})
		if serr != nil {
			return ui.View{}, serr
		}
		sigView = ui.SignatureView{
			Header: manifest.Signing.Header,
			Scheme: manifest.Signing.Scheme + "-" + manifest.Signing.Algorithm,
		}
	}

	view := ui.View{
		Provider:  in.Provider,
		Event:     in.Event,
		Signed:    signed,
		NoSign:    in.NoSign,
		DryRun:    in.DryRun,
		Signature: sigView,
		Request: ui.RequestView{
			Method:  method,
			URL:     in.URL,
			Headers: displayHeaders(headers, sigHeaders),
			Bytes:   len(body),
		},
		Overrides: len(in.Sets),
	}
	if in.DryRun {
		return view, nil
	}

	req, err := fire.BuildRequest(fire.RequestSpec{
		Method:          method,
		URL:             in.URL,
		Body:            body,
		Headers:         headers,
		SignatureHeaders: headerToMap(sigHeaders),
	})
	if err != nil {
		return view, err
	}
	res, err := deps.Sender.Send(context.Background(), req)
	if err != nil {
		view.Err = err.Error()
		return view, &transportError{err}
	}
	view.Response = &ui.ResponseView{
		Status:    res.Status,
		LatencyMS: res.LatencyMS,
		Bytes:     res.Bytes,
		Body:      string(res.Body),
	}
	if in.Fail && (res.Status < 200 || res.Status >= 300) {
		return view, &failResponseError{status: res.Status}
	}
	return view, nil
}

// displayHeaders merges request headers with signature header values for the UI.
func displayHeaders(headers map[string]string, sig http.Header) map[string]string {
	out := make(map[string]string, len(headers)+len(sig))
	for k, v := range headers {
		out[k] = v
	}
	for k, vals := range sig {
		if len(vals) > 0 {
			out[k] = vals[0]
		}
	}
	return out
}

// headerToMap converts http.Header to the map[string][]string expected by fire.RequestSpec.
func headerToMap(h http.Header) map[string][]string {
	out := make(map[string][]string, len(h))
	for k, v := range h {
		out[k] = v
	}
	return out
}
