package handler

import (
	"bytes"
	"compress/flate"
	"compress/gzip"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/nikitaw13/metricscollector/internal/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// testHashKey is the HMAC key shared by the hash middleware tests.
const testHashKey = "TestKey"

// fixedResponseHandler returns a handler replying with the given Content-Type and body.
func fixedResponseHandler(contentType, body string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", contentType)
		w.Write([]byte(body))
	})
}

// TestCompressMiddlewareGzipResponse verifies that responses are gzip-compressed for gzip-capable clients.
func TestCompressMiddlewareGzipResponse(t *testing.T) {
	t.Parallel()
	handler := compressMiddleware(fixedResponseHandler("application/json", `{"status":"ok"}`))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Accept-Encoding", "gzip")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, "gzip", rec.Header().Get("Content-Encoding"))

	gzipReader, err := gzip.NewReader(rec.Body)
	require.NoError(t, err)
	defer gzipReader.Close()

	body, err := io.ReadAll(gzipReader)
	require.NoError(t, err)
	assert.Equal(t, `{"status":"ok"}`, string(body))
}

// TestCompressMiddlewareDeflateResponse verifies deflate compression when only deflate is accepted.
func TestCompressMiddlewareDeflateResponse(t *testing.T) {
	t.Parallel()
	handler := compressMiddleware(fixedResponseHandler("application/json", `{"status":"ok"}`))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Accept-Encoding", "deflate")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, "deflate", rec.Header().Get("Content-Encoding"))

	flateReader := flate.NewReader(rec.Body)
	defer flateReader.Close()

	body, err := io.ReadAll(flateReader)
	require.NoError(t, err)
	assert.Equal(t, `{"status":"ok"}`, string(body))
}

// TestCompressMiddlewareNoCompressionWithoutAcceptEncoding verifies the uncompressed passthrough when no encoding is accepted.
func TestCompressMiddlewareNoCompressionWithoutAcceptEncoding(t *testing.T) {
	t.Parallel()
	handler := compressMiddleware(fixedResponseHandler("application/json", `{"status":"ok"}`))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Empty(t, rec.Header().Get("Content-Encoding"))
	assert.Equal(t, `{"status":"ok"}`, rec.Body.String())
}

// TestCompressMiddlewareMultipleAcceptEncoding verifies that gzip wins when several encodings are accepted.
func TestCompressMiddlewareMultipleAcceptEncoding(t *testing.T) {
	t.Parallel()
	handler := compressMiddleware(fixedResponseHandler("application/json", `{"status":"ok"}`))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Accept-Encoding", "gzip, deflate, br")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, "gzip", rec.Header().Get("Content-Encoding"))
}

// TestCompressMiddlewareSkipsNonCompressibleType verifies that text/plain responses bypass compression.
func TestCompressMiddlewareSkipsNonCompressibleType(t *testing.T) {
	t.Parallel()
	handler := compressMiddleware(fixedResponseHandler("text/plain", "hello"))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Accept-Encoding", "gzip")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Empty(t, rec.Header().Get("Content-Encoding"))
	assert.Equal(t, "hello", rec.Body.String())
}

// TestCompressMiddlewareCompressesHTML verifies that text/html responses are gzip-compressed.
func TestCompressMiddlewareCompressesHTML(t *testing.T) {
	t.Parallel()
	handler := compressMiddleware(fixedResponseHandler("text/html; charset=utf-8", "<html>hi</html>"))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Accept-Encoding", "gzip")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, "gzip", rec.Header().Get("Content-Encoding"))

	gzipReader, err := gzip.NewReader(rec.Body)
	require.NoError(t, err)
	defer gzipReader.Close()

	body, err := io.ReadAll(gzipReader)
	require.NoError(t, err)
	assert.Equal(t, "<html>hi</html>", string(body))
}

// TestDecompressMiddlewareGzipBody verifies that gzip request bodies are decompressed before reaching the handler.
func TestDecompressMiddlewareGzipBody(t *testing.T) {
	t.Parallel()
	var buf bytes.Buffer
	gzipWriter := gzip.NewWriter(&buf)
	_, err := gzipWriter.Write([]byte(`{"status":"ok"}`))
	require.NoError(t, err)
	require.NoError(t, gzipWriter.Close())

	var gotBody string
	handler := decompressMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		b, _ := io.ReadAll(r.Body)
		gotBody = string(b)
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodPost, "/", &buf)
	req.Header.Set("Content-Encoding", "gzip")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, `{"status":"ok"}`, gotBody)
	assert.Empty(t, req.Header.Get("Content-Encoding"))
}

// TestDecompressMiddlewareDeflateBody verifies that deflate request bodies are decompressed before reaching the handler.
func TestDecompressMiddlewareDeflateBody(t *testing.T) {
	t.Parallel()
	var buf bytes.Buffer
	flateWriter, _ := flate.NewWriter(&buf, flate.BestCompression)
	_, err := flateWriter.Write([]byte(`{"status":"ok"}`))
	require.NoError(t, err)
	require.NoError(t, flateWriter.Close())

	var gotBody string
	handler := decompressMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		b, _ := io.ReadAll(r.Body)
		gotBody = string(b)
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodPost, "/", &buf)
	req.Header.Set("Content-Encoding", "deflate")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, `{"status":"ok"}`, gotBody)
	assert.Empty(t, req.Header.Get("Content-Encoding"))
}

// TestDecompressMiddlewareNoEncodingPassthrough verifies that raw bodies pass through unchanged.
func TestDecompressMiddlewareNoEncodingPassthrough(t *testing.T) {
	t.Parallel()
	var gotBody string
	handler := decompressMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		b, _ := io.ReadAll(r.Body)
		gotBody = string(b)
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewBufferString("raw body"))
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, "raw body", gotBody)
}

// TestDecompressMiddlewareInvalidGzipBody verifies that a broken gzip body is rejected with 500.
func TestDecompressMiddlewareInvalidGzipBody(t *testing.T) {
	t.Parallel()
	handler := decompressMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewBufferString("not gzip"))
	req.Header.Set("Content-Encoding", "gzip")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusInternalServerError, rec.Code)
}

// TestRequireJSONContentEmptyBody verifies that an empty JSON request body is rejected with 400.
func TestRequireJSONContentEmptyBody(t *testing.T) {
	t.Parallel()
	handler := requireJSONContent(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(""))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

// TestValidateHashEmptyHeader verifies that a missing HashSHA256 header is rejected with 400.
func TestValidateHashEmptyHeader(t *testing.T) {
	t.Parallel()
	mh := &MetricsHandler{hashKey: testHashKey}
	handler := mh.validateHashMiddleware(fixedResponseHandler("application/json; charset=utf-8", ""))
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("HashSHA256", "")
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

// TestValidateHashCorrectHeaderValue verifies that a valid hash lets the request through with its body intact.
func TestValidateHashCorrectHeaderValue(t *testing.T) {
	t.Parallel()
	mh := &MetricsHandler{hashKey: testHashKey}

	var gotBody []byte
	var err error

	inner := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotBody, err = io.ReadAll(r.Body)
		if err != nil {
			t.Fatal("read body:", err)
		}
	})

	handler := mh.validateHashMiddleware(inner)

	body := `{"id":"test","type":"gauge","value":1}`
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))

	hasher := hmac.New(sha256.New, []byte(mh.hashKey))
	hasher.Write([]byte(body))
	calculatedHash := hasher.Sum(nil)
	calculatedHashHex := hex.EncodeToString(calculatedHash)

	req.Header.Set("HashSHA256", calculatedHashHex)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusOK, rec.Code)

	assert.Equal(t, body, string(gotBody))
}

// TestValidateHashWrongHeaderValue verifies that a mismatching HashSHA256 header is rejected with 400.
func TestValidateHashWrongHeaderValue(t *testing.T) {
	t.Parallel()
	mh := &MetricsHandler{hashKey: testHashKey}
	handler := mh.validateHashMiddleware(fixedResponseHandler("application/json; charset=utf-8", ""))
	body := `{"id":"test","type":"gauge","value":1}`
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))

	hasher := hmac.New(sha256.New, []byte("WRONG VALUE"))
	hasher.Write([]byte(body))
	calculatedHash := hasher.Sum(nil)
	calculatedHashHex := hex.EncodeToString(calculatedHash)

	req.Header.Set("HashSHA256", calculatedHashHex)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

// TestWriteHashHeaderOK verifies that a successful response is signed and delivered unchanged.
func TestWriteHashHeaderOK(t *testing.T) {
	t.Parallel()
	mh := &MetricsHandler{hashKey: testHashKey}
	body := `{"id":"test","type":"gauge","value":1}`
	handler := mh.writeHashHeaderMiddleware(fixedResponseHandler("application/json; charset=utf-8", body))
	req := httptest.NewRequest(http.MethodPost, "/", nil)

	hasher := hmac.New(sha256.New, []byte(mh.hashKey))
	hasher.Write([]byte(body))
	calculatedHash := hasher.Sum(nil)
	calculatedHashHex := hex.EncodeToString(calculatedHash)

	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, body, rec.Body.String())
	assert.Equal(t, calculatedHashHex, rec.Header().Get("HashSHA256"))
}

// TestWriteHashHeaderNonOKStatus verifies that a non-200 response is still signed and its status preserved.
func TestWriteHashHeaderNonOKStatus(t *testing.T) {
	t.Parallel()
	mh := &MetricsHandler{hashKey: testHashKey}
	body := `{"id":"test","type":"gauge","value":1}`
	inner := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusTeapot)
		w.Write([]byte(body))
	})

	handler := mh.writeHashHeaderMiddleware(inner)

	hasher := hmac.New(sha256.New, []byte(mh.hashKey))
	hasher.Write([]byte(body))
	calculatedHash := hasher.Sum(nil)
	calculatedHashHex := hex.EncodeToString(calculatedHash)

	req := httptest.NewRequest(http.MethodPost, "/", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusTeapot, rec.Code)
	assert.Equal(t, body, rec.Body.String())
	assert.Equal(t, calculatedHashHex, rec.Header().Get("HashSHA256"))
}

// TestIntegrationCompressedRequest verifies the request side of the signing contract: a gzip-compressed batch with a valid HashSHA256 header is accepted with 200.
func TestIntegrationCompressedRequest(t *testing.T) {
	t.Parallel()
	ts := GetTestServerWithKey(testHashKey)
	defer ts.Close()

	body := []byte(`[{"id":"test","type":"gauge","value":1.5}]`)

	var buf bytes.Buffer
	gz, err := gzip.NewWriterLevel(&buf, gzip.BestCompression)
	if err != nil {
		t.Error("error while gzip.NewWriterLevel")
		return
	}
	gz.Write(body)
	gz.Close()

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
	ts := GetTestServerWithKey(testHashKey)
	defer ts.Close()

	body := []byte(`{"id":"___test___","type":"gauge"}`)

	var buf bytes.Buffer
	gz, err := gzip.NewWriterLevel(&buf, gzip.BestCompression)
	require.NoError(t, err)
	_, err = gz.Write(body)
	require.NoError(t, err)
	require.NoError(t, gz.Close())

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
