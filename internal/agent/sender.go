package agent

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/nikitaw13/metricscollector/internal/model"
	"github.com/nikitaw13/metricscollector/internal/sign"
	"go.uber.org/zap"
)

type httpClient interface {
	Do(req *http.Request) (*http.Response, error)
}

// Sender is responsible for sending collected metrics to the server
// as a single batched HTTP POST request.
type Sender struct {
	baseURL string
	client  httpClient
	hashKey string
	logger  *zap.Logger
}

// NewSender creates a Sender with the given base URL, HTTP client, and signing key.
func NewSender(baseURL string, client httpClient, hashKey string, logger *zap.Logger) *Sender {
	return &Sender{
		baseURL: baseURL,
		client:  client,
		hashKey: hashKey,
		logger:  logger,
	}
}

// SendBatch sends the given metric batch to the server as a single HTTP POST request.
func (s *Sender) SendBatch(metrics []model.Metric) {
	if len(metrics) == 0 {
		s.logger.Debug("empty metrics batch, skipping send")
		return
	}

	updatesURL := fmt.Sprintf("%s/updates", s.baseURL)

	jsonBody, err := json.Marshal(&metrics)
	if err != nil {
		s.logger.Debug("Failed to marshal a []model.Metric", zap.Error(err))
		return
	}

	compressedBody, err := Compress(jsonBody)
	if err != nil {
		s.logger.Debug("Failed to compress a json body", zap.Error(err))
		return
	}
	req, err := http.NewRequest(
		http.MethodPost,
		updatesURL,
		bytes.NewReader(compressedBody),
	)
	if err != nil {
		s.logger.Debug("Failed to prepare a HTTP request", zap.Error(err))
		return
	}

	if s.hashKey != "" {
		req.Header.Set("HashSHA256", sign.HMAC(s.hashKey, jsonBody))
	}
	req.Header.Set("Content-Type", "application/json; charset=utf-8")
	req.Header.Set("Content-Encoding", "gzip")
	resp, err := s.client.Do(req)
	if err != nil {
		s.logger.Error("Failed to send a HTTP request", zap.Error(err))
		return
	}

	logRequest(s.logger, resp, updatesURL)
	drainAndCloseResponse(resp)
}

// logRequest logs the outcome of sending a metrics batch.
func logRequest(logger *zap.Logger, resp *http.Response, url string) {
	if resp.StatusCode < 300 {
		logger.Debug("Metrics sent", zap.String("url", url))
	} else {
		logger.Error("Metrics failed to send", zap.String("url", url), zap.Int("status", resp.StatusCode))
	}
}

// drainAndCloseResponse drains and closes the response body to allow TCP connection reuse.
func drainAndCloseResponse(resp *http.Response) {
	io.Copy(io.Discard, resp.Body)
	resp.Body.Close()
}
