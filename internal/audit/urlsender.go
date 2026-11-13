package audit

import (
	"bytes"
	"fmt"
	"net/http"
)

// urlSender POSTs events to a URL.
type urlSender struct {
	url    string
	client httpClient
}

// newURLSender creates a urlSender that POSTs events to url via client.
func newURLSender(url string, client httpClient) *urlSender {
	return &urlSender{
		url:    url,
		client: client,
	}
}

// close is a no-op: urlSender holds no resources to release.
func (us *urlSender) close() error {
	return nil
}

// send POSTs event to the configured URL and reports non-2xx responses as errors.
func (us *urlSender) send(event []byte) error {
	req, err := http.NewRequest(http.MethodPost, us.url, bytes.NewReader(event))
	if err != nil {
		return err
	}

	resp, err := us.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= http.StatusInternalServerError {
		return fmt.Errorf("server error while sending event: status code %d", resp.StatusCode)
	}

	if resp.StatusCode >= http.StatusBadRequest {
		return fmt.Errorf("client error while sending event: status code %d", resp.StatusCode)
	}

	if resp.StatusCode >= http.StatusMultipleChoices {
		return fmt.Errorf("unexpected redirect: status code %d", resp.StatusCode)
	}

	return nil
}
