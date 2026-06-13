package ui

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sort"
)

const (
	separator  = " ────────────────────────────────────────────"
	snippetMax = 80
)

func formatBytes(n int) string {
	if n >= 1024 {
		return fmt.Sprintf("%.1fkb", float64(n)/1024)
	}
	return fmt.Sprintf("%db", n)
}

func stripScheme(rawURL string) string {
	u, err := url.Parse(rawURL)
	if err != nil || u.Host == "" {
		return rawURL
	}
	out := u.Host + u.Path
	if u.RawQuery != "" {
		out += "?" + u.RawQuery
	}
	return out
}

// truncate shortens s to at most limit runes (not bytes), so a multi-byte rune
// is never split into invalid UTF-8.
func truncate(s string, limit int) string {
	r := []rune(s)
	if len(r) <= limit {
		return s
	}
	return string(r[:limit-1]) + "…"
}

// RenderHUD writes the colored pipeline view. Deterministic for a given View.
func RenderHUD(w io.Writer, v View) error {
	out := &ew{w: w}

	title := styleDim.Render("hookfire") + "  " + styleBoldCyan.Render(v.Provider+"/"+v.Event)
	out.println(title)
	out.println("")

	arrow := styleAccent.Render("▸")

	renderLine := fmt.Sprintf(" %s %s %s",
		styleLabel.Render("render"),
		arrow,
		styleValue.Render(formatBytes(v.Request.Bytes)+" payload · "+fmt.Sprintf("%d overrides", v.Overrides)),
	)
	out.println(renderLine)

	var signLine string
	if v.NoSign {
		signLine = fmt.Sprintf(" %s %s %s",
			styleLabel.Render("sign  "),
			arrow,
			styleDim.Render("skipped (--no-sign)"),
		)
	} else {
		signLine = fmt.Sprintf(" %s %s %s",
			styleLabel.Render("sign  "),
			arrow,
			styleValue.Render(v.Signature.Scheme+" → "+v.Signature.Header),
		)
	}
	out.println(signLine)

	if v.DryRun {
		fireLine := fmt.Sprintf(" %s %s %s",
			styleLabel.Render("fire  "),
			arrow,
			styleDim.Render("dry-run (not sent)"),
		)
		out.println(fireLine)
		out.println(styleSep.Render(separator))

		keys := make([]string, 0, len(v.Request.Headers))
		for k := range v.Request.Headers {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		for _, k := range keys {
			out.printf(" %s: %s\n", styleLabel.Render(k), styleValue.Render(v.Request.Headers[k]))
		}
		return out.err
	}

	fireLine := fmt.Sprintf(" %s %s %s",
		styleLabel.Render("fire  "),
		arrow,
		styleValue.Render(v.Request.Method+" "+stripScheme(v.Request.URL)),
	)
	out.println(fireLine)
	out.println(styleSep.Render(separator))

	if v.Err != "" {
		out.printf(" %s %s\n", styleErr.Render("✗"), styleDim.Render(v.Err))
		return out.err
	}

	if v.Response != nil {
		status := fmt.Sprintf("%d %s", v.Response.Status, http.StatusText(v.Response.Status))
		meta := fmt.Sprintf("%s · %dms · %s", status, v.Response.LatencyMS, formatBytes(v.Response.Bytes))
		snippet := truncate(v.Response.Body, snippetMax)
		resultLine := fmt.Sprintf(" %s %s   %s",
			styleOK.Render("✓"),
			styleValue.Render(meta),
			styleSnippet.Render(snippet),
		)
		out.println(resultLine)
	}
	return out.err
}
