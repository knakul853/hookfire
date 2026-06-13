package render

import "strconv"

// Vars holds the injected values for built-in {{var}} tokens. Callers fill it
// (timestamp/uuid/now from the clock+rng at the CLI boundary) so render stays
// pure and deterministic in tests.
type Vars struct {
	Timestamp int64
	UUID      string
	Event     string
	Now       string // RFC3339
}

// Map returns the {{var}} substitution table. iso8601 aliases now.
func (v Vars) Map() map[string]string {
	return map[string]string{
		"timestamp": strconv.FormatInt(v.Timestamp, 10),
		"uuid":      v.UUID,
		"event":     v.Event,
		"now":       v.Now,
		"iso8601":   v.Now,
	}
}
