package ui

import "io"

// View is the renderer-agnostic model of a fired (or dry-run) event.
type View struct {
	Provider  string
	Event     string
	Signed    bool
	NoSign    bool
	DryRun    bool
	Signature SignatureView
	Request   RequestView
	Response  *ResponseView // nil on dry-run or transport error
	Err       string        // non-empty on transport error
	Overrides int
}

// SignatureView describes the signature applied (or skipped).
type SignatureView struct {
	Header string
	Scheme string // e.g. "hmac-sha256"
}

// RequestView is the outbound request summary.
type RequestView struct {
	Method  string
	URL     string
	Headers map[string]string
	Bytes   int
}

// ResponseView is the captured response summary.
type ResponseView struct {
	Status    int
	LatencyMS int64
	Bytes     int
	Body      string
}

// Render writes v to w using the selected mode.
func Render(w io.Writer, mode Mode, v View) error {
	switch mode {
	case ModeJSON:
		return RenderJSON(w, v)
	case ModeHUD:
		return RenderHUD(w, v)
	default:
		return RenderPlain(w, v)
	}
}
