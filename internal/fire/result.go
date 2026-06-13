package fire

// Result is the captured outcome of firing a request.
type Result struct {
	Status    int
	LatencyMS int64
	Bytes     int
	Body      []byte
	Headers   map[string][]string
}
