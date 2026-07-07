package event

// Payload is the JSON body POSTed by language SDKs to the ingest endpoint.
type Payload struct {
	Service     string   `json:"service"`
	Environment string   `json:"environment"`
	ErrorType   string   `json:"error_type"`
	Message     string   `json:"message"`
	StackTrace  []string `json:"stack_trace"`
	Timestamp   string   `json:"timestamp"`
	SDKVersion  string   `json:"sdk_version"`
}
