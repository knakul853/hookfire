package render

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"
)

// Render interpolates {{vars}} in tmpl, parses the result as JSON (numbers kept
// as json.Number for byte-faithful round-trips), applies each --set override in
// order, and returns canonical JSON bytes. Those exact bytes are what the caller
// signs and sends.
func Render(tmpl []byte, vars Vars, sets []Set) ([]byte, error) {
	interpolated := interpolate(string(tmpl), vars.Map())

	dec := json.NewDecoder(bytes.NewReader([]byte(interpolated)))
	dec.UseNumber()
	var doc any
	if err := dec.Decode(&doc); err != nil {
		return nil, fmt.Errorf("render: template is not valid JSON after interpolation: %w", err)
	}
	for _, s := range sets {
		updated, err := applySet(doc, s)
		if err != nil {
			return nil, err
		}
		doc = updated
	}
	out, err := json.Marshal(doc)
	if err != nil {
		return nil, fmt.Errorf("render: marshal canonical body: %w", err)
	}
	return out, nil
}

func interpolate(s string, vars map[string]string) string {
	for k, v := range vars {
		s = strings.ReplaceAll(s, "{{"+k+"}}", v)
	}
	return s
}
