package provider

import (
	"embed"
	"io/fs"
)

//go:embed all:builtin
var builtinFS embed.FS

// Builtin returns the embedded built-in providers rooted at the provider dirs.
func Builtin() fs.FS {
	sub, err := fs.Sub(builtinFS, "builtin")
	if err != nil {
		panic(err) // embed path is a compile-time constant; this cannot happen
	}
	return sub
}
