package audit

import (
	"bytes"
	"fmt"
	"net/http"
	"os"
)

const fileMode = os.FileMode(0644)

// fileWriter appends events to a file.
type fileWriter struct {
	file string
}

func (fw fileWriter) send(event []byte) error {
	file, err := os.OpenFile(fw.file, os.O_CREATE|os.O_APPEND|os.O_WRONLY, fileMode)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = fmt.Fprintf(file, "%s\n", event)
	if err != nil {
		return err
	}
	return nil
}

// urlSender POSTs events to a URL.
type urlSender struct {
	url    string
	client httpClient
}

func (us urlSender) send(event []byte) error {
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
