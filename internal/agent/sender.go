package agent

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"

	"github.com/nikitaw13/metricscollector/internal/model"
	"github.com/nikitaw13/metricscollector/internal/sign"
)

// Sender is responsible for sending collected metrics to the server
// as a single batched HTTP POST request.
type Sender struct {
	baseURL string
	client  HTTPClient
	hashKey string
}

// NewSender creates a Sender with the given base URL, HTTP client, and signing key.
func NewSender(baseURL string, client HTTPClient, hashKey string) *Sender {
	return &Sender{
		baseURL: baseURL,
		client:  client,
		hashKey: hashKey,
	}
}

// SendBatch sends the given metric batch to the server as a single HTTP POST request.
func (s *Sender) SendBatch(metrics []model.Metric) {
	if len(metrics) == 0 {
		log.Println("empty metrics batch, skipping send")
		return
	}

	updatesURL := fmt.Sprintf("%s/updates", s.baseURL)

	jsonBody, err := json.Marshal(&metrics)
	if err != nil {
		log.Println(err)
		return
	}

	compressedBody, err := Compress(jsonBody)
	if err != nil {
		log.Println(err)
		return
	}
	req, err := http.NewRequest(
		http.MethodPost,
		updatesURL,
		bytes.NewReader(compressedBody),
	)
	if err != nil {
		log.Println(err)
		return
	}

	if s.hashKey != "" {
		req.Header.Set("HashSHA256", sign.HMAC(s.hashKey, jsonBody))
	}
	req.Header.Set("Content-Type", "application/json; charset=utf-8")
	req.Header.Set("Content-Encoding", "gzip")
	resp, err := s.client.Do(req)
	if err != nil {
		log.Println(err)
		return
	}

	logRequest(resp, updatesURL)
	drainAndCloseResponse(resp)
}

// logRequest logs the outcome of sending a metrics batch.
func logRequest(resp *http.Response, url string) {
	if resp.StatusCode < 300 {
		log.Printf("Metrics sent to %s", url)
	} else {
		log.Printf("Metrics failed to send to %s, status: %d", url, resp.StatusCode)
	}
}

// drainAndCloseResponse drains and closes the response body to allow TCP connection reuse.
func drainAndCloseResponse(resp *http.Response) {
	io.Copy(io.Discard, resp.Body)
	resp.Body.Close()
}
