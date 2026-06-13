package render

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
)

// Set is one --set override: a dot-path and its raw string value. ForceString
// (from --set-string) skips type inference and stores the value as a string.
type Set struct {
	Path        string
	Value       string
	ForceString bool
}

// applySet navigates doc by s.Path (dot-separated; numeric segments index
// arrays) and assigns the inferred value. Missing object segments are created;
// an out-of-range array index is an error.
func applySet(doc any, s Set) (any, error) {
	segs := strings.Split(s.Path, ".")
	val := inferValue(s.Value, s.ForceString)
	updated, err := setPath(doc, segs, val)
	if err != nil {
		return nil, fmt.Errorf("render: --set %s: %w", s.Path, err)
	}
	return updated, nil
}

func setPath(node any, segs []string, val any) (any, error) {
	if len(segs) == 0 {
		return val, nil
	}
	seg := segs[0]
	rest := segs[1:]

	if idx, err := strconv.Atoi(seg); err == nil {
		arr, ok := node.([]any)
		if !ok {
			return nil, fmt.Errorf("segment %q indexes a non-array", seg)
		}
		if idx < 0 || idx >= len(arr) {
			return nil, fmt.Errorf("array index %d out of range (len %d)", idx, len(arr))
		}
		child, err := setPath(arr[idx], rest, val)
		if err != nil {
			return nil, err
		}
		arr[idx] = child
		return arr, nil
	}

	obj, ok := node.(map[string]any)
	if !ok {
		if node == nil {
			obj = map[string]any{}
		} else {
			return nil, fmt.Errorf("segment %q indexes a non-object", seg)
		}
	}
	child, err := setPath(obj[seg], rest, val)
	if err != nil {
		return nil, err
	}
	obj[seg] = child
	return obj, nil
}

// inferValue maps a raw --set string to a JSON value: true/false→bool, null→nil,
// numeric→json.Number (preserving integer form), {…}/[…]→parsed JSON, else
// string. ForceString always yields a string. The numeric branch is gated on
// JSON's own number grammar (not strconv.ParseFloat) so inputs like "007", "Inf"
// or "NaN" — which ParseFloat accepts but JSON rejects — fall through to string
// instead of producing a body that fails to marshal later.
func inferValue(raw string, forceString bool) any {
	if forceString {
		return raw
	}
	switch raw {
	case "true":
		return true
	case "false":
		return false
	case "null":
		return nil
	}
	var n json.Number
	if json.Unmarshal([]byte(raw), &n) == nil {
		return n
	}
	if len(raw) > 0 && (raw[0] == '{' || raw[0] == '[') {
		var parsed any
		if err := json.Unmarshal([]byte(raw), &parsed); err == nil {
			return parsed
		}
	}
	return raw
}
