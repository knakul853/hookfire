// Package cli wires the hookfire cobra commands: trigger, replay, show, list,
// verify, and version. It owns flag parsing, slog verbosity, and exit-code
// mapping; command bodies delegate all real work to the internal packages.
package cli
