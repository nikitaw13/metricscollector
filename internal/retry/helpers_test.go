package retry

import (
	"bytes"
	"net/http"
	"sync"
	"testing"
)

var _ HTTPClient = (*fakeHTTPClient)(nil)

type fakeHTTPClient struct {
	errors []error
	status int

	mu    sync.Mutex
	calls int
}

func (c *fakeHTTPClient) Do(req *http.Request) (*http.Response, error) {
	c.mu.Lock()
	n := c.calls
	c.calls++
	c.mu.Unlock()

	if n < len(c.errors) {
		return nil, c.errors[n]
	}
	status := c.status
	if status == 0 {
		status = http.StatusOK
	}
	return &http.Response{StatusCode: status, Body: http.NoBody}, nil
}

func newRetryRequest(t *testing.T) *http.Request {
	t.Helper()
	req, err := http.NewRequest(
		http.MethodPost,
		"http://retry.test/updates",
		bytes.NewReader([]byte("{}")),
	)
	if err != nil {
		t.Fatalf("error building retry request = %v", err)
	}
	return req
}

func doRequest(t *testing.T, client HTTPClient, req *http.Request) (int, error) {
	t.Helper()

	resp, err := client.Do(req)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	return resp.StatusCode, nil
}
