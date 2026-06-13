package provider

import (
	"errors"
	"fmt"
	"io/fs"
	"sort"
)

// ErrUnknownProvider / ErrUnknownEvent are sentinels for catalog misses; the
// CLI maps them to exit code 2 and suggests near matches.
var (
	ErrUnknownProvider = errors.New("provider: unknown provider")
	ErrUnknownEvent    = errors.New("provider: unknown event")
)

// Catalog resolves a (provider, event) pair to its manifest and event template.
type Catalog interface {
	Lookup(provider, event string) (Manifest, Event, error)
	List() []string
	Events(provider string) ([]string, error)
	// Manifest returns the provider manifest without requiring an event name.
	// Used by replay, which works from a raw payload rather than a named event.
	Manifest(provider string) (Manifest, error)
}

// Sources describes where providers come from, in increasing precedence:
// Embedded built-ins, then each fs.FS in Dirs (left→right). OnShadow is called
// when a later source overrides an already-loaded provider by name.
type Sources struct {
	Embedded fs.FS
	Dirs     []fs.FS
	OnShadow func(name, source string)
}

type catalog struct {
	providers map[string]Provider
}

// NewCatalog loads every provider from sources, later sources overriding earlier
// ones by name. A shadowing override invokes Sources.OnShadow (a warn hook).
func NewCatalog(s Sources) (Catalog, error) {
	c := &catalog{providers: map[string]Provider{}}
	var layers []fs.FS
	if s.Embedded != nil {
		layers = append(layers, s.Embedded)
	}
	layers = append(layers, s.Dirs...)

	for _, layer := range layers {
		names, err := providerDirs(layer)
		if err != nil {
			return nil, err
		}
		for _, name := range names {
			p, err := loadProvider(layer, name)
			if err != nil {
				return nil, err
			}
			if _, exists := c.providers[name]; exists && s.OnShadow != nil {
				s.OnShadow(name, p.Source)
			}
			c.providers[name] = p
		}
	}
	return c, nil
}

func providerDirs(fsys fs.FS) ([]string, error) {
	entries, err := fs.ReadDir(fsys, ".")
	if err != nil {
		return nil, fmt.Errorf("provider: scan dir: %w", err)
	}
	var names []string
	for _, e := range entries {
		if e.IsDir() {
			names = append(names, e.Name())
		}
	}
	return names, nil
}

func (c *catalog) Lookup(provider, event string) (Manifest, Event, error) {
	p, ok := c.providers[provider]
	if !ok {
		return Manifest{}, Event{}, fmt.Errorf("%w: %q", ErrUnknownProvider, provider)
	}
	ev, ok := p.Events[event]
	if !ok {
		return Manifest{}, Event{}, fmt.Errorf("%w: %q for provider %q", ErrUnknownEvent, event, provider)
	}
	return p.Manifest, ev, nil
}

func (c *catalog) List() []string {
	out := make([]string, 0, len(c.providers))
	for name := range c.providers {
		out = append(out, name)
	}
	sort.Strings(out)
	return out
}

func (c *catalog) Manifest(name string) (Manifest, error) {
	p, ok := c.providers[name]
	if !ok {
		return Manifest{}, fmt.Errorf("%w: %q", ErrUnknownProvider, name)
	}
	return p.Manifest, nil
}

func (c *catalog) Events(provider string) ([]string, error) {
	p, ok := c.providers[provider]
	if !ok {
		return nil, fmt.Errorf("%w: %q", ErrUnknownProvider, provider)
	}
	out := make([]string, 0, len(p.Events))
	for name := range p.Events {
		out = append(out, name)
	}
	sort.Strings(out)
	return out, nil
}
