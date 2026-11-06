package audit

import "net/http"

//go:generate mockgen -source=interfaces.go -destination=interfaces_mock_test.go -package=audit

// subscriber consumes serialized audit events.
type subscriber interface {
	send(event []byte) error
}

// httpClient abstracts a single-request HTTP client; *http.Client implements it.
type httpClient interface {
	Do(req *http.Request) (*http.Response, error)
}
