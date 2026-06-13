package ui

import (
	"encoding/json"
	"fmt"
	"io"
)

type jsonOut struct {
	Provider  string        `json:"provider"`
	Event     string        `json:"event"`
	Signed    bool          `json:"signed"`
	Signature jsonSignature `json:"signature"`
	Request   jsonRequest   `json:"request"`
	Response  *jsonResponse `json:"response"`
	DryRun    bool          `json:"dry_run"`
}

type jsonSignature struct {
	Header string `json:"header"`
	Scheme string `json:"scheme"`
}

type jsonRequest struct {
	Method  string            `json:"method"`
	URL     string            `json:"url"`
	Headers map[string]string `json:"headers"`
	Bytes   int               `json:"bytes"`
}

type jsonResponse struct {
	Status    int    `json:"status"`
	LatencyMS int64  `json:"latency_ms"`
	Bytes     int    `json:"bytes"`
	Body      string `json:"body"`
}

// RenderJSON writes the locked machine-readable representation of v.
func RenderJSON(w io.Writer, v View) error {
	out := jsonOut{
		Provider: v.Provider, Event: v.Event, Signed: v.Signed,
		Signature: jsonSignature{Header: v.Signature.Header, Scheme: v.Signature.Scheme},
		Request: jsonRequest{Method: v.Request.Method, URL: v.Request.URL,
			Headers: v.Request.Headers, Bytes: v.Request.Bytes},
		DryRun: v.DryRun,
	}
	if v.Response != nil {
		out.Response = &jsonResponse{
			Status: v.Response.Status, LatencyMS: v.Response.LatencyMS,
			Bytes: v.Response.Bytes, Body: v.Response.Body,
		}
	}
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	if err := enc.Encode(out); err != nil {
		return fmt.Errorf("ui: encode json: %w", err)
	}
	return nil
}
