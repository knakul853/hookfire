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

// fireParams is the resolved input to the shared sign→fire→view path.
type fireParams struct {
	manifest     provider.Manifest
	provider     string
	event        string
	url          string
	body         []byte
	secret       config.Secret
	timestamp    int64
	now          func() int64
	baseHeaders  map[string]string
	extraHeaders map[string]string
	overrides    int
	noSign       bool
	dryRun       bool
	fail         bool
}

// fireResult signs body (unless noSign or scheme==none), assembles the ui.View,
// and (unless dryRun) builds and sends the request via deps.Sender. The view is
// populated even on transport error / --fail non-2xx so callers can render it.
func fireResult(deps pipelineDeps, p fireParams) (ui.View, error) {
	now := p.now
	if now == nil {
		now = realNow
	}
	ts := p.timestamp
	if ts == 0 {
		ts = now()
	}

	method := p.manifest.Transport.Method
	if method == "" {
		method = http.MethodPost
	}
	headers := make(map[string]string, len(p.baseHeaders)+len(p.extraHeaders))
	for k, v := range p.baseHeaders {
		headers[k] = v
	}
	for k, v := range p.extraHeaders {
		headers[k] = v
	}

	signed := !p.noSign && p.manifest.Signing.Scheme != "none"
	sigHeaders := http.Header{}
	var sigView ui.SignatureView
	if signed {
		signer, err := sign.New(p.manifest.SigningConfig())
		if err != nil {
			return ui.View{}, err
		}
		if p.secret.Reveal() == "" {
			return ui.View{}, sign.ErrMissingSecret
		}
		var serr error
		sigHeaders, serr = signer.Sign(p.body, sign.Options{Secret: p.secret.Reveal(), Timestamp: ts})
		if serr != nil {
			return ui.View{}, serr
		}
		sigView = ui.SignatureView{
			Header: p.manifest.Signing.Header,
			Scheme: p.manifest.Signing.Scheme + "-" + p.manifest.Signing.Algorithm,
		}
	}

	view := ui.View{
		Provider:  p.provider,
		Event:     p.event,
		Signed:    signed,
		NoSign:    p.noSign,
		DryRun:    p.dryRun,
		Signature: sigView,
		Request: ui.RequestView{
			Method:  method,
			URL:     p.url,
			Headers: displayHeaders(headers, sigHeaders),
			Bytes:   len(p.body),
		},
		Overrides: p.overrides,
	}
	if p.dryRun {
		return view, nil
	}

	req, err := fire.BuildRequest(fire.RequestSpec{
		Method: method, URL: p.url, Body: p.body,
		Headers: headers, SignatureHeaders: sigHeaders,
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
		Status: res.Status, LatencyMS: res.LatencyMS, Bytes: res.Bytes, Body: string(res.Body),
	}
	if p.fail && (res.Status < 200 || res.Status >= 300) {
		return view, &failResponseError{status: res.Status}
	}
	return view, nil
}

// runPipeline resolves the event, renders the body, and delegates to fireResult.
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
	body, err := render.Render(event.Template, render.Vars{
		Timestamp: ts, UUID: genUUID(), Event: in.Event,
		Now: time.Unix(ts, 0).UTC().Format(time.RFC3339),
	}, in.Sets)
	if err != nil {
		return ui.View{}, err
	}
	return fireResult(deps, fireParams{
		manifest: manifest, provider: in.Provider, event: in.Event, url: in.URL,
		body: body, secret: in.Secret, timestamp: ts, now: now,
		baseHeaders: event.Headers(), extraHeaders: in.ExtraHeaders,
		overrides: len(in.Sets), noSign: in.NoSign, dryRun: in.DryRun, fail: in.Fail,
	})
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
