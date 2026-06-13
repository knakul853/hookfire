package ui

import (
	"fmt"
	"io"
)

// ew wraps an io.Writer and accumulates the first write error, silencing
// subsequent writes. It lets a renderer make a run of Fprintf calls and check
// the error once at the end, without unchecked-error lint noise on each.
type ew struct {
	w   io.Writer
	err error
}

func (e *ew) printf(format string, args ...any) {
	if e.err != nil {
		return
	}
	_, e.err = fmt.Fprintf(e.w, format, args...)
}

func (e *ew) println(s string) {
	if e.err != nil {
		return
	}
	_, e.err = fmt.Fprintln(e.w, s)
}
