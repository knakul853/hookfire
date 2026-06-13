package provider

import (
	"fmt"
	"io/fs"
	"path"
	"strings"
)

// Provider is a fully loaded provider: a validated manifest and its events.
type Provider struct {
	Manifest Manifest
	Events   map[string]Event
	Source   string // where it came from (for logging shadow warnings)
}

// loadProvider reads <root>/manifest.yaml and every <root>/events/*.json from
// fsys, validating the manifest fail-closed before returning.
func loadProvider(fsys fs.FS, root string) (Provider, error) {
	raw, err := fs.ReadFile(fsys, path.Join(root, "manifest.yaml"))
	if err != nil {
		return Provider{}, fmt.Errorf("provider %q: read manifest: %w", root, err)
	}
	m, err := ParseManifest(raw)
	if err != nil {
		return Provider{}, err
	}
	if err := m.Validate(); err != nil {
		return Provider{}, err
	}

	events := map[string]Event{}
	eventsDir := path.Join(root, "events")
	entries, err := fs.ReadDir(fsys, eventsDir)
	if err == nil {
		for _, e := range entries {
			if e.IsDir() || !strings.HasSuffix(e.Name(), ".json") {
				continue
			}
			data, rerr := fs.ReadFile(fsys, path.Join(eventsDir, e.Name()))
			if rerr != nil {
				return Provider{}, fmt.Errorf("provider %q: read event %s: %w", m.Name, e.Name(), rerr)
			}
			name := strings.TrimSuffix(e.Name(), ".json")
			mc := m
			events[name] = Event{Name: name, Template: data, manifest: &mc}
		}
	}
	return Provider{Manifest: m, Events: events, Source: root}, nil
}
