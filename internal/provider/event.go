package provider

// Event is one resolved provider event: its name, its raw JSON template bytes,
// and a back-reference to its manifest for transport/per-event headers.
type Event struct {
	Name     string
	Template []byte
	manifest *Manifest
}

// Headers returns the transport headers plus per-event overrides plus
// Content-Type. Per-event headers win over transport headers.
func (e Event) Headers() map[string]string {
	out := map[string]string{}
	if e.manifest != nil {
		if ct := e.manifest.Transport.ContentType; ct != "" {
			out["Content-Type"] = ct
		}
		for k, v := range e.manifest.Transport.Headers {
			out[k] = v
		}
		if em, ok := e.manifest.Events[e.Name]; ok {
			for k, v := range em.Headers {
				out[k] = v
			}
		}
	}
	return out
}
