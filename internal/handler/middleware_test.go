package handler

import (
	"bytes"
	"compress/flate"
	"compress/gzip"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// okHandler returns an http.Handler that responds with the given content type and body.
func okHandler(contentType, body string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", contentType)
		w.Write([]byte(body))
	})
}

func TestCompressMiddleware_GzipResponse(t *testing.T) {
	handler := CompressMiddleware(okHandler("application/json", `{"status":"ok"}`))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Accept-Encoding", "gzip")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, "gzip", rec.Header().Get("Content-Encoding"))

	gr, err := gzip.NewReader(rec.Body)
	require.NoError(t, err)
	defer gr.Close()

	body, err := io.ReadAll(gr)
	require.NoError(t, err)
	assert.Equal(t, `{"status":"ok"}`, string(body))
}

func TestCompressMiddleware_DeflateResponse(t *testing.T) {
	handler := CompressMiddleware(okHandler("application/json", `{"status":"ok"}`))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Accept-Encoding", "deflate")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, "deflate", rec.Header().Get("Content-Encoding"))

	fr := flate.NewReader(rec.Body)
	defer fr.Close()

	body, err := io.ReadAll(fr)
	require.NoError(t, err)
	assert.Equal(t, `{"status":"ok"}`, string(body))
}

func TestCompressMiddleware_NoCompressionWithoutAcceptEncoding(t *testing.T) {
	handler := CompressMiddleware(okHandler("application/json", `{"status":"ok"}`))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Empty(t, rec.Header().Get("Content-Encoding"))
	assert.Equal(t, `{"status":"ok"}`, rec.Body.String())
}

func TestCompressMiddleware_MultipleAcceptEncoding(t *testing.T) {
	handler := CompressMiddleware(okHandler("application/json", `{"status":"ok"}`))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Accept-Encoding", "gzip, deflate, br")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, "gzip", rec.Header().Get("Content-Encoding"))
}

func TestCompressMiddleware_SkipsNonCompressibleType(t *testing.T) {
	handler := CompressMiddleware(okHandler("text/plain", "hello"))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Accept-Encoding", "gzip")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Empty(t, rec.Header().Get("Content-Encoding"))
	assert.Equal(t, "hello", rec.Body.String())
}

func TestCompressMiddleware_CompressesHTML(t *testing.T) {
	handler := CompressMiddleware(okHandler("text/html; charset=utf-8", "<html>hi</html>"))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Accept-Encoding", "gzip")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, "gzip", rec.Header().Get("Content-Encoding"))

	gr, err := gzip.NewReader(rec.Body)
	require.NoError(t, err)
	defer gr.Close()

	body, err := io.ReadAll(gr)
	require.NoError(t, err)
	assert.Equal(t, "<html>hi</html>", string(body))
}

func TestDecompressMiddleware_GzipBody(t *testing.T) {
	var buf bytes.Buffer
	gw := gzip.NewWriter(&buf)
	_, err := gw.Write([]byte(`{"status":"ok"}`))
	require.NoError(t, err)
	require.NoError(t, gw.Close())

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

func TestDecompressMiddleware_DeflateBody(t *testing.T) {
	var buf bytes.Buffer
	fw, _ := flate.NewWriter(&buf, flate.BestCompression)
	_, err := fw.Write([]byte(`{"status":"ok"}`))
	require.NoError(t, err)
	require.NoError(t, fw.Close())

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

func TestDecompressMiddleware_NoEncodingPassthrough(t *testing.T) {
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

func TestDecompressMiddleware_InvalidGzipBody(t *testing.T) {
	handler := DecompressMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewBufferString("not gzip"))
	req.Header.Set("Content-Encoding", "gzip")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusInternalServerError, rec.Code)
}
