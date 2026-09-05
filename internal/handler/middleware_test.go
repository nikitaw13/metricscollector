package handler

import (
	"bytes"
	"compress/flate"
	"compress/gzip"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func fixedResponseHandler(contentType, body string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", contentType)
		w.Write([]byte(body))
	})
}

func TestCompressMiddlewareGzipResponse(t *testing.T) {
	handler := CompressMiddleware(fixedResponseHandler("application/json", `{"status":"ok"}`))

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

func TestCompressMiddlewareDeflateResponse(t *testing.T) {
	handler := CompressMiddleware(fixedResponseHandler("application/json", `{"status":"ok"}`))

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

func TestCompressMiddlewareNoCompressionWithoutAcceptEncoding(t *testing.T) {
	handler := CompressMiddleware(fixedResponseHandler("application/json", `{"status":"ok"}`))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Empty(t, rec.Header().Get("Content-Encoding"))
	assert.Equal(t, `{"status":"ok"}`, rec.Body.String())
}

func TestCompressMiddlewareMultipleAcceptEncoding(t *testing.T) {
	handler := CompressMiddleware(fixedResponseHandler("application/json", `{"status":"ok"}`))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Accept-Encoding", "gzip, deflate, br")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, "gzip", rec.Header().Get("Content-Encoding"))
}

func TestCompressMiddlewareSkipsNonCompressibleType(t *testing.T) {
	handler := CompressMiddleware(fixedResponseHandler("text/plain", "hello"))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Accept-Encoding", "gzip")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Empty(t, rec.Header().Get("Content-Encoding"))
	assert.Equal(t, "hello", rec.Body.String())
}

func TestCompressMiddlewareCompressesHTML(t *testing.T) {
	handler := CompressMiddleware(fixedResponseHandler("text/html; charset=utf-8", "<html>hi</html>"))

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

func TestDecompressMiddlewareGzipBody(t *testing.T) {
	var buf bytes.Buffer
	gzipWriter := gzip.NewWriter(&buf)
	_, err := gzipWriter.Write([]byte(`{"status":"ok"}`))
	require.NoError(t, err)
	require.NoError(t, gzipWriter.Close())

	var gotBody string
	handler := DecompressMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
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

func TestDecompressMiddlewareDeflateBody(t *testing.T) {
	var buf bytes.Buffer
	flateWriter, _ := flate.NewWriter(&buf, flate.BestCompression)
	_, err := flateWriter.Write([]byte(`{"status":"ok"}`))
	require.NoError(t, err)
	require.NoError(t, flateWriter.Close())

	var gotBody string
	handler := DecompressMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
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

func TestDecompressMiddlewareNoEncodingPassthrough(t *testing.T) {
	var gotBody string
	handler := DecompressMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
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

func TestDecompressMiddlewareInvalidGzipBody(t *testing.T) {
	handler := DecompressMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewBufferString("not gzip"))
	req.Header.Set("Content-Encoding", "gzip")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusInternalServerError, rec.Code)
}

func TestRequireJSONContentEmptyBody(t *testing.T) {
	handler := requireJSONContent(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(""))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
}
