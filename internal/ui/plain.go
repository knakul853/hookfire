package ui

import "io"

// RenderPlain writes aligned key: value lines with no ANSI — the off-TTY view.
func RenderPlain(w io.Writer, v View) error {
	out := &ew{w: w}
	out.printf("hookfire  %s/%s\n", v.Provider, v.Event)
	out.printf("render:  %d bytes · %d overrides\n", v.Request.Bytes, v.Overrides)
	switch {
	case v.NoSign:
		out.printf("sign:    skipped (--no-sign)\n")
	default:
		out.printf("sign:    %s → %s\n", v.Signature.Scheme, v.Signature.Header)
	}
	if v.DryRun {
		out.printf("fire:    dry-run (not sent)\n")
		return out.err
	}
	if v.Err != "" {
		out.printf("fire:    %s\n", v.Err)
		return out.err
	}
	out.printf("fire:    %s %s\n", v.Request.Method, v.Request.URL)
	if v.Response != nil {
		out.printf("result:  %d · %dms · %db\n", v.Response.Status, v.Response.LatencyMS, v.Response.Bytes)
	}
	return out.err
}
