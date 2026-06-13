// Command hookfire fires correctly-shaped, correctly-signed synthetic webhook
// events at a URL. See https://github.com/knakul853/hookfire.
package main

import (
	"os"

	"github.com/knakul853/hookfire/internal/cli"
)

func main() {
	os.Exit(cli.Run(os.Args[1:]))
}
