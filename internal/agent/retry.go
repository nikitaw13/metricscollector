package agent

import (
	"log"
	"net/http"
	"slices"
	"time"
)

// HTTPClient abstracts a single-request HTTP client; *http.Client implements it.
type HTTPClient interface {
	Do(req *http.Request) (*http.Response, error)
}

var _ HTTPClient = (*ClientWithRetries)(nil)

// ClientWithRetries wraps an HTTPClient and retries transport failures with the configured backoff.
type ClientWithRetries struct {
	retryDelays []time.Duration
	client      HTTPClient
}

// NewClientWithRetries wraps client with the given retryDelays between attempts.
func NewClientWithRetries(retryDelays []time.Duration, client HTTPClient) *ClientWithRetries {
	return &ClientWithRetries{
		client:      client,
		retryDelays: slices.Clone(retryDelays),
	}
}

// Do sends the request, retrying transport failures with the configured backoff between attempts.
func (c *ClientWithRetries) Do(req *http.Request) (*http.Response, error) {
	resp, err := c.client.Do(req)

	if err != nil {
		if req.GetBody == nil {
			return nil, err
		}

		log.Println(err)

		for i := 0; i < len(c.retryDelays); i++ {
			time.Sleep(c.retryDelays[i])

			req.Body, err = req.GetBody()

			if err != nil {
				return nil, err
			}
			resp, err = c.client.Do(req)
			if err != nil {
				log.Printf("retry %d/%d failed: %v", i+1, len(c.retryDelays), err)
				continue
			}
			break
		}
		if err != nil {
			return nil, err
		}
	}
	return resp, nil
}
