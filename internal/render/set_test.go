package render

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSetTypeInference(t *testing.T) {
	cases := []struct {
		in, path, val, want string
		str                 bool
	}{
		{`{"a":1}`, "a", "5", `{"a":5}`, false},
		{`{"a":1}`, "a", "true", `{"a":true}`, false},
		{`{"a":1}`, "a", "null", `{"a":null}`, false},
		{`{"a":1}`, "a", `{"k":1}`, `{"a":{"k":1}}`, false},
		{`{"a":1}`, "a", "hello", `{"a":"hello"}`, false},
		{`{"a":1}`, "a", "5", `{"a":"5"}`, true},
	}
	for _, c := range cases {
		out, err := Render([]byte(c.in), Vars{}, []Set{{Path: c.path, Value: c.val, ForceString: c.str}})
		require.NoError(t, err)
		require.JSONEq(t, c.want, string(out))
	}
}

func TestSetNestedCreateAndArrayIndex(t *testing.T) {
	out, err := Render([]byte(`{"deployment":{"target":"prod"},"items":[{"n":1}]}`), Vars{}, []Set{
		{Path: "deployment.target", Value: "null"},
		{Path: "items.0.n", Value: "9"},
		{Path: "meta.new.flag", Value: "true"},
	})
	require.NoError(t, err)
	require.JSONEq(t, `{"deployment":{"target":null},"items":[{"n":9}],"meta":{"new":{"flag":true}}}`, string(out))
}

func TestSetArrayIndexOutOfRangeErrors(t *testing.T) {
	_, err := Render([]byte(`{"items":[{"n":1}]}`), Vars{}, []Set{{Path: "items.5.n", Value: "9"}})
	require.Error(t, err)
}

func TestSetIntegerStaysInteger(t *testing.T) {
	out, err := Render([]byte(`{"amount":0}`), Vars{}, []Set{{Path: "amount", Value: "2000"}})
	require.NoError(t, err)
	require.Equal(t, `{"amount":2000}`, string(out))
}

// Values that strconv.ParseFloat accepts but JSON rejects (leading zeros, Inf,
// NaN) must become strings rather than a body that fails to marshal.
func TestSetNonJSONNumberBecomesString(t *testing.T) {
	for _, raw := range []string{"007", "Inf", "NaN"} {
		out, err := Render([]byte(`{"v":0}`), Vars{}, []Set{{Path: "v", Value: raw}})
		require.NoError(t, err, "value %q", raw)
		require.JSONEq(t, `{"v":"`+raw+`"}`, string(out), "value %q", raw)
	}
}
