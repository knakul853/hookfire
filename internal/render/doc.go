// Package render turns an event template into the exact bytes hookfire signs and
// sends. It interpolates {{vars}} (timestamp, uuid, event, now, iso8601), parses
// the result as JSON (numbers preserved as json.Number for byte-faithful
// round-trips), applies --set dot-path overrides with type inference, and
// marshals canonical bytes. Dynamic values are injected from outside so renders
// are deterministic and testable.
package render
