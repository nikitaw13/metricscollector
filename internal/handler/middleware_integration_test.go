package handler

import (
	"bytes"
	"compress/gzip"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"io"
	"net/http"
	"testing"

	"github.com/nikitaw13/metricscollector/internal/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestIntegrationCompressedRequest verifies the request side of the signing contract: a gzip-compressed batch with a valid HashSHA256 header is accepted with 200.
func TestIntegrationCompressedRequest(t *testing.T) {
	t.Parallel()
	ts := newTestServerWithKey(testHashKey)
	defer ts.Close()

	body := []byte(`[{"id":"test","type":"gauge","value":1.5}]`)

	var buf bytes.Buffer
	gzipWriter, err := gzip.NewWriterLevel(&buf, gzip.BestCompression)
	if err != nil {
		t.Error("error while gzip.NewWriterLevel")
		return
	}
	gzipWriter.Write(body)
	gzipWriter.Close()

	hasher := hmac.New(sha256.New, []byte(testHashKey))
	hasher.Write(body)
	calculatedHash := hasher.Sum(nil)
	calculatedHashHex := hex.EncodeToString(calculatedHash)
	req, err := http.NewRequest(http.MethodPost, ts.URL+"/updates", &buf)
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Content-Encoding", "gzip")
	req.Header.Set("HashSHA256", calculatedHashHex)

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

// TestIntegrationCompressedResponse verifies the response side of the signing
// contract: with a key configured and a gzip-capable client, the response
// arrives gzip-compressed while HashSHA256 still carries the HMAC of the
// uncompressed body.
func TestIntegrationCompressedResponse(t *testing.T) {
	t.Parallel()
	ts := newTestServerWithKey(testHashKey)
	defer ts.Close()

	body := []byte(`{"id":"___test___","type":"gauge"}`)

	var buf bytes.Buffer
	gzipWriter, err := gzip.NewWriterLevel(&buf, gzip.BestCompression)
	require.NoError(t, err)
	_, err = gzipWriter.Write(body)
	require.NoError(t, err)
	require.NoError(t, gzipWriter.Close())

	hasher := hmac.New(sha256.New, []byte(testHashKey))
	hasher.Write(body)
	req, err := http.NewRequest(http.MethodPost, ts.URL+"/value", &buf)
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Content-Encoding", "gzip")
	req.Header.Set("HashSHA256", hex.EncodeToString(hasher.Sum(nil)))
	req.Header.Set("Accept-Encoding", "gzip")

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Equal(t, "gzip", resp.Header.Get("Content-Encoding"))

	gzipReader, err := gzip.NewReader(resp.Body)
	require.NoError(t, err)
	defer gzipReader.Close()

	plainResp, err := io.ReadAll(gzipReader)
	require.NoError(t, err)

	var got model.Metric
	require.NoError(t, json.Unmarshal(plainResp, &got))
	assert.Equal(t, "___test___", got.ID)
	require.NotNil(t, got.Value)
	assert.Equal(t, defaultGaugeValue, *got.Value)

	responseHasher := hmac.New(sha256.New, []byte(testHashKey))
	responseHasher.Write(plainResp)
	assert.Equal(t, hex.EncodeToString(responseHasher.Sum(nil)), resp.Header.Get("HashSHA256"),
		"response hash must be computed over the uncompressed body")
}
