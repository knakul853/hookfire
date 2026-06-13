package fire

// Result is the captured outcome of firing a request.
type Result struct {
	Status    int
	LatencyMS int64 // total time from send to fully-read body, not time-to-first-byte
	Bytes     int
	Body      []byte
	Headers   map[string][]string
}
