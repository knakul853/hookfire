// Package fire sends a prepared HTTP request to the target and captures the
// response (status, latency, bytes, body) into a Result. The Sender interface is
// mocked so CLI tests never hit the network; the http implementation supports an
// --insecure TLS-skip option (which the CLI warns about).
package fire
