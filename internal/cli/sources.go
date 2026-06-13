package cli

import (
	"io/fs"
	"os"
	"path/filepath"

	"github.com/knakul853/hookfire/internal/provider"
)

// providerDirPaths returns the existing filesystem provider directories in
// increasing precedence: $XDG_CONFIG_HOME/hookfire/providers, ./providers, then
// the --providers-dir flag. Nonexistent directories are skipped.
func providerDirPaths(xdgConfigHome, cwd, flagDir string) []string {
	var paths []string
	if xdgConfigHome != "" {
		if p := filepath.Join(xdgConfigHome, "hookfire", "providers"); isDir(p) {
			paths = append(paths, p)
		}
	}
	if cwd != "" {
		if p := filepath.Join(cwd, "providers"); isDir(p) {
			paths = append(paths, p)
		}
	}
	if flagDir != "" && isDir(flagDir) {
		paths = append(paths, flagDir)
	}
	return paths
}

func isDir(p string) bool {
	info, err := os.Stat(p)
	return err == nil && info.IsDir()
}

// buildCatalog assembles the catalog from the embedded built-ins plus the
// filesystem provider directories (later overriding earlier by name). onShadow
// is invoked when a filesystem provider shadows an already-loaded one.
func buildCatalog(flagDir string, onShadow func(name, source string)) (provider.Catalog, error) {
	cwd, err := os.Getwd()
	if err != nil {
		cwd = ""
	}
	paths := providerDirPaths(os.Getenv("XDG_CONFIG_HOME"), cwd, flagDir)
	dirs := make([]fs.FS, 0, len(paths))
	for _, p := range paths {
		dirs = append(dirs, os.DirFS(p))
	}
	return provider.NewCatalog(provider.Sources{
		Embedded: provider.Builtin(),
		Dirs:     dirs,
		OnShadow: onShadow,
	})
}
